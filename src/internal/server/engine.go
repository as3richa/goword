package server

import (
  "time"
  "net/http"

	"internal/log"

  "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	writeWait = 10 * time.Second
	pongWait = 30 * time.Second
	pingPeriod = 25 * time.Second
	maxMessageSize = 1024
)

func engineHandler(w http.ResponseWriter, r *http.Request) {
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
}
