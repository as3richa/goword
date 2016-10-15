package engine

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"internal/grid"
	"internal/log"
)

const (
	betweenGameDuration = 1 * time.Minute
	countdownDuration   = 5 * time.Second
	gameDuration        = 3 * time.Minute
)

const (
	stateAwaitingPlayers = "awaitingPlayers"
	stateBetweenGames    = "betweenGames"
	stateCountdown       = "countdown"
	stateInGame          = "inGame"
)

var wordRegex = regexp.MustCompile("^[a-zA-Z]+$")

type lobby struct {
	Name  string `json:"name"`
	State string `json:"state"`

	asyncInterrupt *time.Timer
	asyncTimestamp time.Time

	terminator         chan struct{}
	incomingPipe       chan incomingMessage
	parentIncomingPipe chan incomingMessage

	Clients clientSet `json:"players"`

	Grid grid.Grid `json:"grid"`
}

type clientSet map[*Client]*clientData

type clientData struct {
	Playing        bool `json:"playing"`
	Readied        bool `json:"readied"`
	Score          int  `json:"score"`
	words          []string
	PreviousResult *clientGameResult `json:"result,omitempty"`
}

type clientGameResult struct {
	Score int          `json:"score"`
	Words []scoredWord `json:"words"`
}

type scoredWord struct {
	Word   string `json:"word"`
	Points int    `json:"points"`
}

var lobbyDispatchTable = [messageTypeCount]func(*lobby, *Client, interface{}){
	lobbyHandleNew,
	lobbyHandleQuit,
	lobbyHandleJoin,
	lobbyHandlePart,
	lobbyHandleReady,
	lobbyHandleWord,
}

func (e *Engine) newLobby(name string) *lobby {
	l := lobby{
		Name:               name,
		State:              stateAwaitingPlayers,
		asyncInterrupt:     time.NewTimer(0),
		asyncTimestamp:     time.Now(),
		terminator:         make(chan struct{}, 1),
		incomingPipe:       newIncomingPipe(),
		parentIncomingPipe: e.incomingPipe,
		Clients:            map[*Client]*clientData{},
		Grid:               grid.Generate(nil),
	}
	l.clearAsyncInterrupt()
	return &l
}

func (l *lobby) run() {
	log.Fields{"lobby": l.Name}.Debug("lobby event loop starting")

	for {
		select {
		case <-l.terminator:
			log.Fields{"lobby": l.Name}.Debug("lobby received termination signal; halting")
			return

		case message := <-l.incomingPipe:
			lobbyDispatchTable[message.what](l, message.client, message.payload)

		case <-l.asyncInterrupt.C:
			if !time.Now().After(l.asyncTimestamp) {
				l.resetAsyncInterrupt(l.asyncTimestamp.Sub(time.Now()))
			}
		}

		l.transitionState()
	}
}

func (l *lobby) terminate() {
	close(l.terminator)
}

func (l *lobby) empty() bool {
	return len(l.Clients) == 0
}

func (l *lobby) readyPlayerCount() int {
	total := 0
	for _, data := range l.Clients {
		if data.Readied {
			total += 1
		}
	}
	return total
}

func (l *lobby) clearAsyncInterrupt() {
	l.asyncInterrupt.Stop()
	select {
	case <-l.asyncInterrupt.C:
	default:
	}
	l.asyncTimestamp = time.Now()
}

func (l *lobby) resetAsyncInterrupt(d time.Duration) {
	l.clearAsyncInterrupt()
	l.asyncInterrupt.Reset(d)
	l.asyncTimestamp = time.Now().Add(d)
}

func (l *lobby) broadcastState(memo string) {
	for client := range l.Clients {
		client.OutgoingPipe <- client.StateMessage(memo)
	}
}

func (l *lobby) transitionState() {
	asyncEvent := time.Now().After(l.asyncTimestamp)

	transition := true
	memo := ""

	switch l.State {
	case stateAwaitingPlayers:
		if len(l.Clients) >= 2 {
			log.Fields{"lobby": l.Name}.Debug("lobby was awaitingPlayers, but now sufficient players are here")
			l.transitionToBetweenGames()
			memo = fmt.Sprintf("Sufficient players; countdown to next game starts in %d seconds", betweenGameDuration/time.Second)
		} else {
			transition = false
		}

	case stateBetweenGames:
		if len(l.Clients) <= 1 {
			log.Fields{"lobby": l.Name}.Debug("lobby was betweenGames, but now insufficient players are here")
			l.transitionToAwaitingPlayers()
			memo = "Insufficient players to start the game; waiting for more..."
		} else if asyncEvent {
			log.Fields{"lobby": l.Name}.Debug("lobby was betweenGames, but the timer has elapsed")
			l.transitionToCountdown()
			memo = fmt.Sprintf("Waiting period is over; game starts in %d seconds", countdownDuration/time.Second)
		} else if l.readyPlayerCount() == len(l.Clients) {
			log.Fields{"lobby": l.Name}.Debug("lobby was betweenGames, but all players have readied up")
			l.transitionToCountdown()
			memo = fmt.Sprintf("Everyone is ready for the next game; game starts in %d seconds", countdownDuration/time.Second)
		} else {
			transition = false
		}

	case stateCountdown:
		if len(l.Clients) <= 1 {
			log.Fields{"lobby": l.Name}.Debug("lobby was in countdown, but now insufficient players are here")
			l.transitionToAwaitingPlayers()
			memo = "Insufficient players to start the game; waiting for more..."
		} else if asyncEvent {
			log.Fields{"lobby": l.Name}.Debug("lobby was in countdown, but the timer has elapsed")
			l.transitionToInGame()
			memo = "Game begin!"
		} else {
			transition = false
		}

	case stateInGame:
		if asyncEvent {
			log.Fields{"lobby": l.Name}.Debug("lobby was inGame, but the timer has elapsed")
			l.endGame()
			if len(l.Clients) <= 1 {
				l.transitionToAwaitingPlayers()
			} else {
				l.transitionToBetweenGames()
			}
			memo = "Game has concluded"
		} else if len(l.Clients) == 0 {
			log.Fields{"lobby": l.Name}.Debug("lobby was inGame, but everyone has left")
			l.transitionToAwaitingPlayers()
		} else {
			transition = false
		}
	}

	if transition {
		l.broadcastState(memo)
	}
}

func (l *lobby) endGame() {
	log.Fields{"lobby": l.Name}.Debug("game is over; scoring")

	playingClientData := make([]*clientData, 0, len(l.Clients))
	wordlists := make([][]string, 0, len(l.Clients))
	for _, data := range l.Clients {
		if !data.Playing {
			data.PreviousResult = nil
		} else {
			playingClientData = append(playingClientData, data)
			wordlists = append(wordlists, data.words)
		}
	}

	log.Fields{"lobby": l.Name, "count": len(playingClientData)}.Debug("playing client count")

	totals, scores := l.Grid.Score(wordlists)
	for i, clientData := range playingClientData {
		clientData.Score += totals[i]
		clientData.PreviousResult = &clientGameResult{
			Score: totals[i],
			Words: make([]scoredWord, len(scores[i])),
		}

		for j, points := range scores[i] {
			clientData.PreviousResult.Words[j] = scoredWord{
				Word:   wordlists[i][j],
				Points: points,
			}
		}
	}
}

func (l *lobby) transitionToAwaitingPlayers() {
	log.Fields{"lobby": l.Name}.Debug("state transition to awaitingPlayers")
	l.clearAsyncInterrupt()
	l.State = stateAwaitingPlayers
	for _, data := range l.Clients {
		data.Playing = true
		data.Readied = false
	}
}

func (l *lobby) transitionToBetweenGames() {
	log.Fields{"lobby": l.Name}.Debug("state transition to betweenGames")
	l.resetAsyncInterrupt(betweenGameDuration)
	l.State = stateBetweenGames
	for _, data := range l.Clients {
		data.Playing = true
		data.Readied = false
	}
}

func (l *lobby) transitionToCountdown() {
	log.Fields{"lobby": l.Name}.Debug("state transition to countdown")
	l.resetAsyncInterrupt(countdownDuration)
	l.State = stateCountdown
	for _, data := range l.Clients {
		data.Readied = false
		data.words = data.words[:0]
		data.PreviousResult = nil
	}
}

func (l *lobby) transitionToInGame() {
	l.resetAsyncInterrupt(gameDuration)
	l.State = stateInGame
	l.Grid = grid.Generate(nil)
	log.Fields{"lobby": l.Name}.Debug("state transition to inGame")
}

func lobbyHandleNew(l *lobby, client *Client, _ interface{}) {
	l.Clients[client] = &clientData{
		Playing: true,
		Readied: false,
		Score:   0,
	}
	l.broadcastState(client.Nickname + " has joined " + l.Name)

	if l.State != stateAwaitingPlayers && l.State != stateBetweenGames {
		client.OutgoingPipe <- client.StateMessage("A game is already in progress; you may join the next round of the game")
	}

	log.Fields{"lobby": l.Name, "client": client.Nickname}.Debug("client joined lobby")
}

func lobbyHandleQuit(l *lobby, client *Client, _ interface{}) {
	lobbyHandlePart(l, client, nil)
	client.Quit()
}

func lobbyHandleJoin(l *lobby, client *Client, _ interface{}) {
	client.OutgoingPipe <- clientErrorMessage{
		Command: "join",
		Message: "You are already in a lobby",
	}

	log.Fields{"lobby": l.Name, "client": client.Nickname}.Debug("client tried to join, but is already in a lobby")
}

func lobbyHandlePart(l *lobby, client *Client, _ interface{}) {
	delete(l.Clients, client)
	client.Lobby = nil
	client.incomingPipe = l.parentIncomingPipe

	client.OutgoingPipe <- client.StateMessage("You have left " + l.Name)
	l.broadcastState(client.Nickname + " has left " + l.Name)

	log.Fields{"lobby": l.Name, "client": client.Nickname}.Debug("client parted lobby")
}

func lobbyHandleReady(l *lobby, client *Client, _ interface{}) {
	if l.State != stateBetweenGames {
		client.OutgoingPipe <- clientErrorMessage{
			Command: "ready",
			Message: "You may only ready up between games",
		}

		log.Fields{"lobby": l.Name, "client": client.Nickname}.Debug("client tried to ready up, but lobby is not between games")
		return
	}

	l.Clients[client].Readied = true
	l.broadcastState(fmt.Sprintf("%s is ready for the next round; %d of %d players are ready", client.Nickname, l.readyPlayerCount(), len(l.Clients)))

	log.Fields{"lobby": l.Name, "client": client.Nickname}.Debug("client has readied up")
}

func lobbyHandleWord(l *lobby, client *Client, data interface{}) {
	word := data.(string)

	if l.State != stateInGame {
		client.OutgoingPipe <- clientErrorMessage{
			Command: "word",
			Message: "You may only record a word during a game",
		}

		log.Fields{"lobby": l.Name, "client": client.Nickname}.Debug("client tried to record a word, but lobby is not in-game")
		return
	}

	if !l.Clients[client].Playing {
		client.OutgoingPipe <- clientErrorMessage{
			Command: "word",
			Message: "You are not playing in the current game",
		}

		log.Fields{"lobby": l.Name, "client": client.Nickname}.Debug("client tried to record a word, but is not currently playing")
		return
	}

	if !wordRegex.MatchString(word) {
		client.OutgoingPipe <- clientErrorMessage{
			Command: "word",
			Message: "You may record only single, non-empty words, containing only letters",
		}

		log.Fields{"lobby": l.Name, "client": client.Nickname}.Debug("client tried to record a word, but word was malformed")
		return
	}

	l.Clients[client].words = append(l.Clients[client].words, word)
	client.OutgoingPipe <- clientWordMessage{
		Word: strings.ToLower(word),
	}
	log.Fields{"lobby": l.Name, "client": client.Nickname}.Debug("client recorded a word")
}
