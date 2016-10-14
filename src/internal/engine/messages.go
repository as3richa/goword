package engine

import (
	"encoding/json"
	"time"
)

const (
	incomingBuffering = 16384
	outgoingBuffering = 128
)

type incomingMessage struct {
	what    incomingMessageType
	client  *Client
	payload interface{}
}

type OutgoingMessage interface{}

func newIncomingPipe() chan incomingMessage {
	return make(chan incomingMessage, incomingBuffering)
}

func newOutgoingPipe() chan OutgoingMessage {
	return make(chan OutgoingMessage, outgoingBuffering)
}

type incomingMessageType int

const (
	messageTypeNew incomingMessageType = iota
	messageTypeQuit
	messageTypeJoin
	messageTypePart
	messageTypeReady
	messageTypeWord
	messageTypeCount
)

type clientStateMessage struct {
	Message string `json:"message,omitempty"`
	*Client
}

type clientErrorMessage struct {
	Command string `json:"command"`
	Message string `json:"message"`
}

type clientWordMessage struct {
	Word string `json:"word"`
}

func (c *Client) StateMessage(memo string) clientStateMessage {
	return clientStateMessage{
		Message: memo,
		Client:  c,
	}
}

func (c clientSet) MarshalJSON() ([]byte, error) {
	result := map[string]*clientData{}
	for client, data := range c {
		result[client.Nickname] = data
	}
	return json.Marshal(result)
}

func (l *lobby) MarshalJSON() ([]byte, error) {
	remaining := (float64)(l.asyncTimestamp.Sub(time.Now())) / (float64)(time.Second)
	ptr := &remaining
	if remaining < 0 {
		ptr = nil
	}

	type Alias lobby
	return json.Marshal(&struct {
		SecondsRemaining *float64 `json:"secondsRemaining,omitempty"`
		*Alias
	}{
		SecondsRemaining: ptr,
		Alias:            (*Alias)(l),
	})
}

func (m clientStateMessage) MarshalJSON() ([]byte, error) {
	type Alias clientStateMessage
	return json.Marshal(&struct {
		Type string `json:"type"`
		Alias
	}{
		Type:  "state",
		Alias: (Alias)(m),
	})
}

func (m clientErrorMessage) MarshalJSON() ([]byte, error) {
	type Alias clientErrorMessage
	return json.Marshal(&struct {
		Type string `json:"type"`
		Alias
	}{
		Type:  "error",
		Alias: (Alias)(m),
	})
}

func (m clientWordMessage) MarshalJSON() ([]byte, error) {
	type Alias clientWordMessage
	return json.Marshal(&struct {
		Type string `json:"type"`
		Alias
	}{
		Type:  "word",
		Alias: (Alias)(m),
	})
}
