package engine

import "internal/log"

const serverBufferSize = 1024

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

		var response Response

		switch message := wrap.Message.(type) {
		case connectMessage:
			response = e.connectClient(client)
		case quitMessage:
			response = e.quitClient(client)
		case joinLobbyMessage:
			response = e.joinClientToLobby(client, message.Name, message.Password, message.Nickname)
		case partLobbyMessage:
			response = e.partClientFromLobby(client)
		default:
			response = badMessageResponse{
				Command: "???",
				Ok:      false,
				Message: "malformed message",
			}
		}

		if response != nil {
			client.SendTo(response)
		}

		if _, ok := response.(quitResponse); ok {
			close(client.ResponsePipe)
		}
	}
}

func (e *Engine) connectClient(client *Client) Response {
	if _, ok := e.clients[client]; ok {
		log.Panic("client connected twice")
	}

	e.clients[client] = struct{}{}
	return connectResponse{
		Command: "connect",
		Ok:      true,
		Message: "you are now connected",
	}
}

func (e *Engine) quitClient(client *Client) Response {
	if !client.alive {
		return nil
	}
	client.alive = false

	if _, ok := e.clients[client]; !ok {
		log.Panic("client quitting, but wasn't connected")
	}
	delete(e.clients, client)

	_ = e.partClientFromLobby(client)

	return quitResponse{
		Command: "quit",
		Ok:      true,
		Message: "goodnight",
	}
}
