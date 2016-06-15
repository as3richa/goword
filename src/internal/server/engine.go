package server

import (
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
	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			c.Quit()
			return
		}
		message, err := engine.UnmarshalMessage(data)
		c.SendFrom(message)
	}
}

func (c *client) Writer() {
	ticker := time.NewTicker(pingPeriod)

	defer ticker.Stop()
	defer c.Close()

	for {
		select {
		case response, ok := <-c.ResponsePipe:
			if !ok {
				_ = c.WriteMessage(websocket.CloseMessage, []byte("bye"))
				return
			}

			data, err := engine.MarshalResponse(response)
			if err != nil {
				log.Fields{"error": err}.Panic("couldn't marshal response")
			}

			c.WriteMessage(websocket.TextMessage, data)
		case <-ticker.C:
			if err := c.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				return
			}
		}
	}
}
