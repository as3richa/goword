package engine

import (
	"encoding/json"
	"fmt"
)

type Message interface{}
type Response interface{}

type wrappedMessage struct {
	*Client
	Message
}

type connectMessage struct{}
type connectResponse struct {
	Command string `json:"command"`
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

type quitMessage struct{}
type quitResponse struct {
	Command string `json:"command"`
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

type joinLobbyMessage struct {
	Name     string
	Password string
	Nickname string
}
type joinLobbyResponse struct {
	Command string   `json:"command"`
	Ok      bool     `json:"ok"`
	Message string   `json:"message,omitempty"`
	Players []string `json:"players,omitempty"`
}

type leaveLobbyMessage struct{}

type badMessageResponse struct {
	Command string `json:"command"`
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

func UnmarshalMessage(data []byte) (Message, error) {
	var object map[string]string
	if err := json.Unmarshal(data, &object); err != nil {
		return nil, err
	}

	switch object["command"] {
	case "join":
		return joinLobbyMessage{
			Name:     object["name"],
			Password: object["password"],
			Nickname: object["nickname"],
		}, nil
	case "quit":
		return quitMessage{}, nil
	case "":
		return nil, fmt.Errorf("missing command parameter")
	default:
		return nil, fmt.Errorf("no such command %s", object["command"])
	}
}

func MarshalResponse(resp Response) ([]byte, error) {
	return json.Marshal(resp)
}
