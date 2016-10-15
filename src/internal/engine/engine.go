package engine

import (
	"regexp"
	"strings"
	"time"

	"internal/log"
	"internal/nickname"
)

const (
	engineHeartbeatInterval = time.Second * 5
	lobbyTimeToLive         = time.Second * 30
)

var lobbyNameRegex = regexp.MustCompile("^[\\w-]+$")

type Engine struct {
	incomingPipe chan incomingMessage
	terminator   chan struct{}

	nicknameGenerator nickname.Generator

	lobbies  map[string]*lobby
	joinedAt map[string]time.Time
}

var engineDispatchTable = [messageTypeCount]func(*Engine, *Client, interface{}){
	engineHandleNew,
	engineHandleQuit,
	engineHandleJoin,
	engineHandlePart,
	engineHandleReady,
	engineHandleWord,
}

func New() *Engine {
	return &Engine{
		incomingPipe:      newIncomingPipe(),
		nicknameGenerator: nickname.Generator{},
		lobbies:           map[string]*lobby{},
		joinedAt:          map[string]time.Time{},
	}
}

func (e *Engine) Run() {
	heartbeat := time.Tick(engineHeartbeatInterval)

	for {
		select {
		case <-e.terminator:
			log.Debug("engine received termination signal; halting engine and all child lobbies")
			for _, lobby := range e.lobbies {
				lobby.terminate()
			}
		case message := <-e.incomingPipe:
			engineDispatchTable[message.what](e, message.client, message.payload)
		case <-heartbeat:
			e.garbageCollectLobbies()
		}
	}
}

func (e *Engine) Terminate() {
	close(e.terminator)
}

func (e *Engine) garbageCollectLobbies() {
	for lobbyName, lobby := range e.lobbies {
		if lobby.empty() {
			if delta := time.Since(e.joinedAt[lobbyName]); delta > lobbyTimeToLive {
				log.Fields{"lobby": lobbyName, "delta": delta}.Info("lobby is empty and was not recently joined; garbage collecting")
				lobby.terminate()
				delete(e.lobbies, lobbyName)
				delete(e.joinedAt, lobbyName)
			}
		}
	}
}

func engineHandleNew(e *Engine, client *Client, _ interface{}) {
	client.Nickname = e.nicknameGenerator.Generate()
	client.OutgoingPipe <- client.StateMessage("Welcome to Goword; you are known as " + client.Nickname)
	log.Fields{"client": client.Nickname}.Debug("new client connected to engine")
}

func engineHandleQuit(e *Engine, client *Client, _ interface{}) {
	e.nicknameGenerator.Free(client.Nickname)
	close(client.OutgoingPipe)
	log.Fields{"client": client.Nickname}.Debug("client quit engine")
}

func engineHandleJoin(e *Engine, client *Client, data interface{}) {
	lobbyName := data.(string)

	if !lobbyNameRegex.MatchString(lobbyName) {
		client.OutgoingPipe <- clientErrorMessage{
			Command: "join",
			Message: "Lobby name may contain only letters, numbers, dashes, and underscores, and may not be empty",
		}
		return
	}

	normalizedName := strings.ToLower(lobbyName)

	var lobby *lobby
	var ok bool

	if lobby, ok = e.lobbies[normalizedName]; !ok {
		log.Fields{"client": client.Nickname, "lobby": lobbyName}.Info("instantiating new lobby")
		lobby = e.newLobby(lobbyName)
		e.lobbies[normalizedName] = lobby
		go lobby.run()
	}

	e.joinedAt[normalizedName] = time.Now()

	client.incomingPipe = lobby.incomingPipe
	client.Lobby = lobby

	client.incomingPipe <- incomingMessage{
		what:   messageTypeNew,
		client: client,
	}

	log.Fields{"client": client.Nickname, "lobby": lobbyName}.Debug("client joining lobby")
}

func engineHandlePart(e *Engine, client *Client, _ interface{}) {
	client.OutgoingPipe <- clientErrorMessage{
		Command: "part",
		Message: "You are not in a lobby",
	}

	log.Fields{"client": client.Nickname}.Debug("client attempted to part lobby, but was not in a lobby")
}

func engineHandleReady(e *Engine, client *Client, _ interface{}) {
	client.OutgoingPipe <- clientErrorMessage{
		Command: "ready",
		Message: "You are not in a lobby",
	}

	log.Fields{"client": client.Nickname}.Debug("client attempted to ready up, but was not in a lobby")
}

func engineHandleWord(e *Engine, client *Client, _ interface{}) {
	client.OutgoingPipe <- clientErrorMessage{
		Command: "word",
		Message: "You are not in a lobby",
	}

	log.Fields{"client": client.Nickname}.Debug("client attempted to guess a word, but was not in a lobby")
}
