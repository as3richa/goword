package engine

import (
	"regexp"
	"strings"
	"time"

	"internal/log"
)

var nicknameRegex = regexp.MustCompile("^[\\w-]+$")
var lobbyNameRegex = nicknameRegex

type lobby struct {
	name     string
	password string

	timerShutdown chan struct{}
	timer         *time.Timer

	nicknames      map[string]struct{}
	clientMetadata map[*Client]struct {
		nickname string
	}
}

func (l *lobby) players() []string {
	names := []string{}
	for name := range l.nicknames {
		names = append(names, name)
	}
	return names
}

func (e *Engine) newLobby(name, password string) *lobby {
	return &lobby{
		name:      name,
		password:  password,
		nicknames: map[string]struct{}{},
	}
}

func (e *Engine) joinClientToLobby(client *Client, name, password, nickname string) Response {
	if client.lobby != nil {
		return joinLobbyResponse{
			Command: "join",
			Ok:      false,
			Message: "you are already in a lobby",
		}
	}

	nickname = strings.TrimSpace(nickname)
	if !nicknameRegex.MatchString(nickname) {
		return joinLobbyResponse{
			Command: "join",
			Ok:      false,
			Message: "nickname cannot be empty, and may contain only letters, numbers, dashes, and underscores",
		}
	}
	normalizedNickname := strings.ToLower(nickname)

	name = strings.TrimSpace(name)
	if !lobbyNameRegex.MatchString(name) {
		return joinLobbyResponse{
			Command: "join",
			Ok:      false,
			Message: "lobby name cannot be empty, and may contain only letters, numbers, dashes, and underscores",
		}
	}
	normalizedName := strings.ToLower(name)

	var lobby = e.lobbies[normalizedName]
	if lobby == nil {
		e.lobbies[normalizedName] = e.newLobby(name, password)
		log.Fields{"name": name}.Info("created new lobby")
		lobby = e.lobbies[normalizedName]
	}

	if lobby.password != password {
		return joinLobbyResponse{
			Command: "join",
			Ok:      false,
			Message: "invalid password",
		}
	}

	if _, ok := lobby.nicknames[normalizedNickname]; ok {
		return joinLobbyResponse{
			Command: "join",
			Ok:      false,
			Message: "nickname is already in use",
		}
	}

	client.lobby = lobby
	lobby.nicknames[normalizedNickname] = struct{}{}

	return joinLobbyResponse{
		Command: "join",
		Ok:      true,
		Message: "you have joined the lobby",
		Players: lobby.players(),
	}
}
