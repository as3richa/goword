package server

import (
	"encoding/json"
	"net/http"
	"time"

	"internal/engine"
	"internal/log"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

const (
	readBufferSize  = 1024
	writeBufferSize = 1024
	writeWait       = 10 * time.Second
	pingPeriod      = 25 * time.Second
	pongWait        = 30 * time.Second
	maxMessageSize  = 1024
)

type client struct {
	*engine.Client
	*websocket.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
}

func engineHandler(engine *engine.Engine) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fields{"error": err}.Info("failed to establish websocket connection")
			return
		}

		client := newClient(engine, ws)
		go client.Writer()
		client.Reader()
	}
}

func newClient(engine *engine.Engine, ws *websocket.Conn) client {
	ws.SetReadLimit(maxMessageSize)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	return client{
		Client: engine.NewClient(),
		Conn:   ws,
	}
}

func (c *client) Reader() {
	defer log.Debug("HTTP engine reader terminating")
	log.Debug("HTTP engine reader running")

	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			c.Quit()
			return
		}

		message := map[string]string{}
		if err = json.Unmarshal(data, &message); err != nil {
			log.Fields{"error": err}.Debug("error unmarshalling incoming JSON payload")
			c.Quit()
			return
		}

		switch message["command"] {
		case "join":
			c.Join(message["lobbyName"])
		case "part":
			c.Part()
		case "ready":
			c.Ready()
		case "word":
			c.Word(message["word"])
		}
	}
}

func (c *client) Writer() {
	ticker := time.NewTicker(pingPeriod)

	defer ticker.Stop()
	defer c.Close()
	defer log.Debug("HTTP engine writer terminating")
	log.Debug("HTTP engine writer running")

	for {
		select {
		case response, ok := <-c.OutgoingPipe:
			if !ok {
				_ = c.WriteMessage(websocket.CloseMessage, []byte("bye"))
				return
			}

			data, err := json.Marshal(response)
			if err != nil {
				log.Fields{"error": err}.Panic("couldn't marshal response")
			}

			if err = c.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				return
			}
		}
	}
}
