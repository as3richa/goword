package engine

import "internal/log"

const clientBufferSize = 16
const serverBufferSize = 1024

type container interface {
	Send(wrappedMessage)
}

type Client struct {
	container container
	Pipe      chan Response
}

type lobby struct {
	name     string
	password string
	clients  map[Client]bool
	pipe     chan wrappedMessage
}

type Engine struct {
	lobbies map[string]lobby
	clients map[*Client]bool
	pipe    chan wrappedMessage
}

func New() *Engine {
	return &Engine{
		lobbies: map[string]lobby{},
		clients: map[*Client]bool{},
		pipe:    make(chan wrappedMessage, serverBufferSize),
	}
}

func (e *Engine) Send(m wrappedMessage) {
	e.pipe <- m
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
			e.clients[client] = true
			client.Send(connectResponse{
				motd: "Hello world!",
			})

		default:
			log.Fields{"message": message}.Panic("engine received unsupported message type")
		}
	}
}
