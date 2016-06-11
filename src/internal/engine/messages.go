package engine

type Message interface{}
type Response interface{}

type wrappedMessage struct {
	*Client
	Message
}

type connectMessage struct{}
type joinLobbyMessage struct {
	name     string
	password string
}
type leaveLobbyMessage struct{}

type connectResponse struct {
	motd string
}
