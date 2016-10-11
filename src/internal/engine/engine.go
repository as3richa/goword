package engine

import (
	"regexp"
	"strings"
	"time"

	"internal/log"
	"internal/nickname"
)

const heartbeatInterval = time.Second * 5

const engineBuffering = 4096
const clientBuffering = 128

var lobbyNameRegex = regexp.MustCompile("^[\\w\\.-]+$")

const lobbyTimeToLive = time.Second * 30

type upstream struct {
	newClientPipe chan *Client
	quitPipe      chan *Client
	joinPipe      chan clientLobbyNamePair
	partPipe      chan *Client
	readyPipe     chan *Client
	guessPipe     chan clientGuessPair
}

type Engine struct {
	upstream
	usedNicknames map[string]struct{}
}

type Client struct {
	ResponsePipe chan Response

	upstream *upstream
	nickname string
}

type Response struct {
	Command  string                 `json:"command"`
	Ok       bool                   `json:"ok"`
	Message  string                 `json:"message,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type clientLobbyNamePair struct {
	client    *Client
	lobbyName string
}

type clientGuessPair struct {
	client *Client
	guess  string
}

func newUpstream() upstream {
	return upstream{
		newClientPipe: make(chan *Client, engineBuffering),
		quitPipe:      make(chan *Client, engineBuffering),
		joinPipe:      make(chan clientLobbyNamePair, engineBuffering),
		partPipe:      make(chan *Client, engineBuffering),
		readyPipe:     make(chan *Client, engineBuffering),
		guessPipe:     make(chan clientGuessPair, engineBuffering),
	}
}

func New() *Engine {
	return &Engine{
		upstream:      newUpstream(),
		usedNicknames: map[string]struct{}{},
	}
}

func (e *Engine) Run() {
	usedNicknames := map[string]struct{}{}

	lobbies := map[string]*lobby{}
	lastLobbyJoin := map[string]time.Time{}

	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case newClient := <-e.newClientPipe:
			var nick string
			for {
				nick = nickname.Generate()
				if _, ok := e.usedNicknames[nick]; !ok {
					break
				}
			}
			e.usedNicknames[nick] = struct{}{}
			newClient.nickname = nick

			newClient.ResponsePipe <- Response{
				Command: "nick",
				Ok:      true,
				Message: "You have connected to the server; you are known as " + nick,
				Metadata: map[string]interface{}{
					"nickname": nick,
				},
			}

		case quittingClient := <-e.quitPipe:
			log.Fields{"client": quittingClient.nickname}.Debug("client quitting")

			delete(usedNicknames, quittingClient.nickname)
			close(quittingClient.ResponsePipe)

		case lobbyNameAndClient := <-e.joinPipe:
			joiningClient := lobbyNameAndClient.client
			lobbyName := lobbyNameAndClient.lobbyName
			normalizedName := strings.ToLower(lobbyName)

			if !lobbyNameRegex.MatchString(lobbyName) {
				joiningClient.ResponsePipe <- Response{
					Command: "join",
					Ok:      false,
					Message: "Lobby name may contain only letters, numbers, dashes, underscores, and periods, and may not be empty",
					Metadata: map[string]interface{}{
						"lobbyName": lobbyName,
					},
				}
			} else {
				var lobby *lobby
				var ok bool

				if lobby, ok = lobbies[normalizedName]; !ok {
					lobby = e.newLobby(lobbyName)
					lobbies[normalizedName] = lobby
					lastLobbyJoin[normalizedName] = time.Now()
					go lobby.run()
				}

				joiningClient.upstream = &lobby.upstream
				lobby.newClientPipe <- joiningClient
			}

		case partingClient := <-e.partPipe:
			partingClient.ResponsePipe <- Response{
				Command: "part",
				Ok:      false,
				Message: "You are not joined to a lobby",
			}

		case readyClient := <-e.readyPipe:
			readyClient.ResponsePipe <- Response{
				Command: "ready",
				Ok:      false,
				Message: "You are not currently playing a game",
			}

		case guessAndClient := <-e.guessPipe:
			guessAndClient.client.ResponsePipe <- Response{
				Command: "guess",
				Ok:      false,
				Message: "You are not currently playing a game",
			}

		case <-heartbeat.C:
			for lobbyName, lobby := range lobbies {
				if lobby.empty() && time.Since(lastLobbyJoin[lobbyName]) > lobbyTimeToLive {
					log.Fields{"lobby": lobbyName, "since": time.Since(lastLobbyJoin[lobbyName])}.Info("lobby is empty and not recently joined; GCing")
					close(lobby.terminator)
					delete(lobbies, lobbyName)
					delete(lastLobbyJoin, lobbyName)
				}
			}
		}
	}
}

func (e *Engine) NewClient() *Client {
	client := &Client{
		ResponsePipe: make(chan Response, clientBuffering),
		upstream:     &e.upstream,
	}

	e.newClientPipe <- client

	return client
}

func (c *Client) Quit() {
	c.upstream.quitPipe <- c
}

func (c *Client) Join(lobbyName string) {
	c.upstream.joinPipe <- clientLobbyNamePair{
		client:    c,
		lobbyName: lobbyName,
	}
}

func (c *Client) Part() {
	c.upstream.partPipe <- c
}

func (c *Client) Ready() {
	c.upstream.readyPipe <- c
}

func (c *Client) Guess(guess string) {
	c.upstream.guessPipe <- clientGuessPair{
		client: c,
		guess:  guess,
	}
}
