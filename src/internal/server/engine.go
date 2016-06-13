package server

import (
	"net/http"
	"time"

	"internal/log"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

const (
	readBufferSize  = 1024
	writeBufferSize = 1024
	writeWait       = 10 * time.Second
	pongWait        = 30 * time.Second
	pingPeriod      = 25 * time.Second
	maxMessageSize  = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
}

func engineHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fields{"error": err}.Info("failed to establish websocket connection")
		return
	}

	defer ws.Close()
	ws.SetReadLimit(maxMessageSize)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	log.Info("client connected")

	for {
		messageType, p, err := ws.ReadMessage()
		if err != nil || string(p) == "bye" {
			break
		}
		if err = ws.WriteMessage(messageType, p); err != nil {
			break
		}
	}
}
