package engine

import (
	"fmt"
	"strings"
	"time"

	"internal/log"
)

const clientBufferSize = 16
const serverBufferSize = 1024

type Client struct {
	alive        bool
	messagePipe  chan wrappedMessage
	ResponsePipe chan Response
}

type lobby struct {
	name     string
	password string

	timerShutdown chan struct{}
	timer         *time.Timer

	clientMetadata map[*Client]struct {
		name string
	}
}

type Engine struct {
	lobbies map[string]*lobby
	clients map[*Client]struct{}
	pipe    chan wrappedMessage
}

func New() *Engine {
	return &Engine{
		lobbies: map[string]*lobby{},
		clients: map[*Client]struct{}{},
		pipe:    make(chan wrappedMessage, serverBufferSize),
	}
}

func (e *Engine) Run() {
	for {
		wrap := <-e.pipe
		client := wrap.Client

		log.Debug("engine received message")

		switch message := wrap.Message.(type) {
		case connectMessage:
			if _, ok := e.clients[client]; ok {
				log.Panic("client connected twice")
			}
			e.clients[client] = struct{}{}
			client.SendTo(connectResponse{
				Command: "connect",
				Ok:      true,
				Motd:    "Hello!",
			})
		case quitMessage:
			if !client.alive {
				return
			}
			client.alive = false
			client.SendTo(quitResponse{
				Command: "quit",
				Ok:      true,
				Message: "Goodnight.",
			})
			close(client.ResponsePipe)
			delete(e.clients, client)
		case joinLobbyMessage:
			message.Name = strings.TrimSpace(message.Name)
			normalizedName := strings.ToLower(message.Name)

			_, ok := e.lobbies[normalizedName]
			if !ok {
				e.lobbies[normalizedName] = e.NewLobby(message.Name, message.Password)
				log.Fields{"name": normalizedName, "password": message.Password}.Info("created lobby")
			}

			if err := fmt.Errorf("unimplemented :)"); err != nil {
				client.SendTo(joinLobbyResponse{
					Command: "join",
					Ok:      false,
					Message: err.Error(),
				})
			} else {
				client.SendTo(joinLobbyResponse{
					Command: "join",
					Ok:      true,
					Message: ":)",
				})
			}
		default:
			client.SendTo(badMessageResponse{
				Command: "???",
				Ok:      false,
				Message: "malformed message",
			})
		}
	}
}
