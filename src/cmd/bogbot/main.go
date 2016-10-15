package main

import (
	"encoding/json"
	"flag"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"internal/grid"
	"internal/log"

	"github.com/gorilla/websocket"
)

var schemeFlag = flag.String("scheme", "ws", "websockt connection scheme")
var addressFlag = flag.String("server", "127.0.0.1:8080", "Goword server address")
var lobbyFlag = flag.String("lobby", "bots", "Goword lobby name")
var aggressionFlag = flag.Int("aggression", 20, "aggression constant for word guessing")

func jsonGet(data interface{}, path ...string) interface{} {
	for _, component := range path {
		data = (data.(map[string]interface{}))[component]
	}
	return data
}

func joinMessage() []byte {
	payload, _ := json.Marshal(map[string]string{
		"command":   "join",
		"lobbyName": *lobbyFlag,
	})
	return payload
}

func readyMessage() []byte {
	payload, _ := json.Marshal(map[string]string{
		"command": "ready",
	})
	return payload
}

func wordMessage(word string) []byte {
	payload, _ := json.Marshal(map[string]string{
		"command": "word",
		"word":    word,
	})
	return payload
}

func main() {
	rand.Seed(time.Now().Unix())

	flag.Parse()

	u := url.URL{Scheme: *schemeFlag, Host: *addressFlag, Path: "/engine"}
	log.Fields{"server": u.String()}.Info("connecting to Goword server")

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {
		log.Fields{"error": err}.Fatal("couldn't connect to Goword server")
	}

	defer func() {
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.Close()
	}()

	wait := make(chan struct{})
	go signalHandler(wait)

	incomingMessages := make(chan interface{}, 100)
	go func() {
		defer close(wait)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Fields{"error": err}.Error("failed to read from websocket client")
				return
			}

			var data interface{}
			if err := json.Unmarshal(message, &data); err != nil {
				log.Fields{"error": err}.Error("failed to unmarshal JSON payload")
				return
			}

			incomingMessages <- data
		}
	}()

	go func() {
		nickname := ""
		state := ""
		haveBoard := false
		var board grid.Grid
		var solution []string

		heartbeat := time.NewTicker(1 * time.Second)

		defer close(wait)
		defer heartbeat.Stop()

		for {
			select {
			case data := <-incomingMessages:
				if messageType := jsonGet(data, "type"); messageType == "state" {
					if len(nickname) == 0 {
						nickname = jsonGet(data, "nickname").(string)
						log.Fields{"nickname": nickname}.Info("received nickname")
					}

					lobby := jsonGet(data, "lobby")

					if lobby == nil {
						err = c.WriteMessage(websocket.TextMessage, joinMessage())
						if err != nil {
							log.Fields{"error": err}.Error("failed to send JSON request")
							return
						}
					} else {
						state = jsonGet(lobby, "state").(string)
						if state == "inGame" {
							if !haveBoard {
								slices := jsonGet(lobby, "grid").([]interface{})
								for i := 0; i < 4; i++ {
									slice := slices[i].([]interface{})
									for j := 0; j < 4; j++ {
										board[i][j] = slice[j].(string)
									}
								}
								solution = board.Solve()
								log.Fields{"board": board, "solution": solution}.Info("Received new grid")
							}
						} else if state == "betweenGames" {
							if readied := jsonGet(lobby, "players", nickname, "readied").(bool); !readied {
								err = c.WriteMessage(websocket.TextMessage, readyMessage())
								if err != nil {
									log.Fields{"error": err}.Error("failed to send JSON request")
									return
								}
							}
						}
						haveBoard = (state == "inGame")
					}
				}
			case <-heartbeat.C:
				if state == "inGame" {
					if rand.Intn(100) < *aggressionFlag {
						err = c.WriteMessage(websocket.TextMessage, wordMessage(solution[rand.Intn(len(solution))]))
						if err != nil {
							log.Fields{"error": err}.Error("failed to send JSON request")
							return
						}
					}
				}
			}
		}
	}()

	<-wait
}

func signalHandler(die chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	for sig := range c {
		log.Fields{"signal": sig}.Info("received signal - terminating")
		close(die)
	}
}
