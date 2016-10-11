package engine

import (
	"fmt"
	"sort"
	"time"

	"internal/grid"
	"internal/log"
)

const forever = time.Hour * 10 * 1000
const betweenGameDuration = 30 * time.Second
const countdownDuration = 5 * time.Second
const gameDuration = 3 * time.Minute

const awaitingPlayers = "awaitingPlayers"
const betweenGames = "betweenGames"
const countdown = "countdown"
const inGame = "inGame"

type lobby struct {
	name string

	async          *time.Timer
	asyncTimestamp time.Time

	terminator chan struct{}
	upstream
	parent *upstream

	clients map[*Client]*clientData
	state   string

	grid grid.Grid
}

type clientData struct {
	playing bool
	readied bool
	guesses []string
}

func (e *Engine) newLobby(name string) *lobby {
	return &lobby{
		name:       name,
		async:      time.NewTimer(forever),
		terminator: make(chan struct{}, 1),
		upstream:   newUpstream(),
		parent:     &e.upstream,
		clients:    map[*Client]*clientData{},
		state:      awaitingPlayers,
	}
}

func (l *lobby) run() {
	for {
		asyncEvent := false

		select {
		case <-l.async.C:
			log.Fields{"lobby": l.name}.Debug("lobby async event fired")
			asyncEvent = true

		case <-l.terminator:
			log.Fields{"lobby": l.name}.Debug("lobby terminating")
			return

		case newClient := <-l.newClientPipe:
			l.clients[newClient] = &clientData{}

			meta := l.metadata()

			newClient.ResponsePipe <- Response{
				Command:  "join",
				Ok:       true,
				Message:  "You have joined " + l.name,
				Metadata: meta,
			}

			message := ""
			if l.state == countdown || l.state == inGame {
				message = "You will join the game in the next round"
			}

			newClient.ResponsePipe <- l.stateResponse(message)

			peerResponse := Response{
				Command:  "join",
				Ok:       true,
				Message:  newClient.nickname + " has joined " + l.name,
				Metadata: meta,
			}

			for client := range l.clients {
				if client != newClient {
					client.ResponsePipe <- peerResponse
				}
			}

		case quittingClient := <-l.quitPipe:
			log.Fields{"lobby": l.name, "client": quittingClient.nickname}.Debug("client quitting")
			l.partClient(quittingClient)
			quittingClient.upstream = l.parent
			quittingClient.Quit()

		case lobbyNameAndClient := <-l.joinPipe:
			joiningClient := lobbyNameAndClient.client

			joiningClient.ResponsePipe <- Response{
				Command: "join",
				Ok:      false,
				Message: "You are already joined to " + l.name,
			}

		case partingClient := <-l.partPipe:
			log.Fields{"lobby": l.name, "client": partingClient.nickname}.Debug("client departing lobby")
			l.partClient(partingClient)

		case readyClient := <-l.readyPipe:
			if l.state != betweenGames {
				log.Fields{"lobby": l.name, "client": readyClient.nickname}.Debug("client readying up, but not between games")

				readyClient.ResponsePipe <- Response{
					Command: "ready",
					Ok:      false,
					Message: "You may only ready up between games",
				}
			} else {
				log.Fields{"lobby": l.name, "client": readyClient.nickname}.Debug("client readying up")

				l.clients[readyClient].readied = true
				if count := l.readyPlayerCount(); count < len(l.clients) {
					response := Response{
						Command: "ready",
						Ok:      true,
						Message: fmt.Sprintf("%d of %d players are ready for the next game", count, len(l.clients)),
					}

					for client := range l.clients {
						client.ResponsePipe <- response
					}
				}
			}

		case guessAndClient := <-l.guessPipe:
			guessingClient := guessAndClient.client
			guess := guessAndClient.guess

			if l.state != inGame {
				log.Fields{"lobby": l.name, "client": guessingClient.nickname}.Debug("client guessing, but not in game")

				guessingClient.ResponsePipe <- Response{
					Command: "ready",
					Ok:      false,
					Message: "You may only guess during a game",
				}
			} else {
				log.Fields{"lobby": l.name, "client": guessingClient.nickname}.Debug("client guessing")

				l.clients[guessingClient].guesses = append(l.clients[guessingClient].guesses, guess)
			}
		}

		l.attemptStateTransition(asyncEvent)
	}
}

func (l *lobby) nicknames() []string {
	nicknames := []string{}
	for client := range l.clients {
		nicknames = append(nicknames, client.nickname)
	}
	sort.Strings(nicknames)
	return nicknames
}

func (l *lobby) empty() bool {
	return len(l.clients) == 0
}

func (l *lobby) partClient(partingClient *Client) {
	delete(l.clients, partingClient)
	partingClient.upstream = l.parent

	partingClient.ResponsePipe <- Response{
		Command: "part",
		Ok:      true,
		Message: "You have left " + l.name,
	}

	meta := l.metadata()

	peerResponse := Response{
		Command:  "part",
		Ok:       true,
		Message:  partingClient.nickname + " has left " + l.name,
		Metadata: meta,
	}

	for client := range l.clients {
		client.ResponsePipe <- peerResponse
	}
}

func (l *lobby) stateResponse(message string) Response {
	return Response{
		Command:  "state",
		Ok:       true,
		Message:  message,
		Metadata: l.metadata(),
	}
}

func (l *lobby) metadata() map[string]interface{} {
	result := map[string]interface{}{}

	result["state"] = l.state
	if remaining := l.asyncTimestamp.Sub(time.Now()); remaining < time.Hour {
		result["remaining"] = (remaining + time.Second - 1) / time.Second
	}

	players := map[string]map[string]interface{}{}
	for client, data := range l.clients {
		players[client.nickname] = map[string]interface{}{
			"playing": data.playing,
			"readied": data.readied,
		}
	}

	result["players"] = players

	if l.state == inGame {
		result["grid"] = l.grid
	}

	return result
}

func (l *lobby) attemptStateTransition(asyncEvent bool) {
	if asyncEvent && !time.Now().After(l.asyncTimestamp) {
		delta := l.asyncTimestamp.Sub(time.Now())
		log.Fields{"delta": delta}.Debug("async event fired early")
		l.resetAsync(delta)
		return
	}

	transition := true
	message := ""

	switch l.state {
	case awaitingPlayers:
		if len(l.clients) >= 2 {
			log.Fields{"lobby": l.name}.Debug("lobby was awaitingPlayers, but now sufficient players are here")
			l.transitionToBetweenGames()
			message = "Sufficient players"
		} else {
			transition = false
		}

	case betweenGames:
		if len(l.clients) <= 1 {
			log.Fields{"lobby": l.name}.Debug("lobby was betweenGames, but now insufficient players are here")
			l.transitionToAwaitingPlayers()
			message = "Insufficient players"
		} else if asyncEvent {
			log.Fields{"lobby": l.name}.Debug("lobby was betweenGames, but the timer has elapsed")
			l.transitionToCountdown()
			message = "Waiting period is over"
		} else if l.readyPlayerCount() == len(l.clients) {
			log.Fields{"lobby": l.name}.Debug("lobby was betweenGames, but all players have readied up")
			l.transitionToCountdown()
			message = "Everyone is ready for the next round"
		} else {
			transition = false
		}

	case countdown:
		if len(l.clients) <= 1 {
			log.Fields{"lobby": l.name}.Debug("lobby was in countdown, but now insufficient players are here")
			l.transitionToAwaitingPlayers()
			message = "Insufficient players"
		} else if asyncEvent {
			log.Fields{"lobby": l.name}.Debug("lobby was in countdown, but the timer has elapsed")
			l.transitionToInGame()
		} else {
			transition = false
		}

	case inGame:
		if asyncEvent {
			log.Fields{"lobby": l.name}.Debug("lobby was inGame, but the timer has elapsed")
			l.endGame()
			if len(l.clients) <= 1 {
				l.transitionToAwaitingPlayers()
			} else {
				l.transitionToBetweenGames()
			}
		} else {
			transition = false
		}
	}

	if transition {
		state := l.stateResponse(message)
		for client := range l.clients {
			client.ResponsePipe <- state
		}
	}
}

func (l *lobby) transitionToAwaitingPlayers() {
	log.Fields{"lobby": l.name}.Debug("state transition to awaitingPlayers")
	l.resetAsync(forever)
	l.state = awaitingPlayers
	for client := range l.clients {
		l.clients[client].playing = true
		l.clients[client].readied = false
	}
}

func (l *lobby) transitionToBetweenGames() {
	log.Fields{"lobby": l.name}.Debug("state transition to betweenGames")
	l.resetAsync(betweenGameDuration)
	l.state = betweenGames
	for client := range l.clients {
		l.clients[client].playing = true
		l.clients[client].readied = false
	}
}

func (l *lobby) transitionToCountdown() {
	log.Fields{"lobby": l.name}.Debug("state transition to countdown")
	l.resetAsync(countdownDuration)
	l.state = countdown
}

func (l *lobby) transitionToInGame() {
	log.Fields{"lobby": l.name}.Debug("state transition to inGame")

	l.resetAsync(gameDuration)
	l.state = inGame

	l.grid = grid.Generate(nil)
	for _, data := range l.clients {
		data.guesses = nil
	}
}

func (l *lobby) endGame() {
	playerNames := []string{}
	lists := [][]string{}

	for client, datum := range l.clients {
		if !datum.playing {
			continue
		}
		playerNames = append(playerNames, client.nickname)
		lists = append(lists, datum.guesses)
	}

	totals, scores := l.grid.Score(lists)

	results := map[string]interface{}{}
	for i, player := range playerNames {
		result := map[string]interface{}{}
		result["total"] = totals[i]

		words := []interface{}{}
		for j, word := range lists[i] {
			words = append(words, []interface{}{word, scores[i][j]})
		}

		result["words"] = words

		results[player] = result
	}

	response := Response{
		Command:  "result",
		Ok:       true,
		Metadata: results,
	}

	for client := range l.clients {
		client.ResponsePipe <- response
	}
}

func (l *lobby) resetAsync(d time.Duration) {
	l.async.Stop()
	select {
	case <-l.async.C:
	default:
	}
	l.asyncTimestamp = time.Now().Add(d)
	l.async.Reset(d + 20*time.Millisecond)
}

func (l *lobby) readyPlayerCount() int {
	total := 0
	for _, datum := range l.clients {
		if datum.readied {
			total += 1
		}
	}
	return total
}
