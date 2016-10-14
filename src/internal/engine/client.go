package engine

type Client struct {
	incomingPipe chan incomingMessage
	OutgoingPipe chan OutgoingMessage `json:"-"`

	Nickname string `json:"nickname"`
	Lobby    *lobby `json:"lobby,omitempty"`
}

func (e *Engine) NewClient() *Client {
	client := &Client{
		incomingPipe: e.incomingPipe,
		OutgoingPipe: newOutgoingPipe(),
	}

	client.incomingPipe <- incomingMessage{
		what:   messageTypeNew,
		client: client,
	}

	return client
}

func (c *Client) Quit() {
	c.incomingPipe <- incomingMessage{
		what:   messageTypeQuit,
		client: c,
	}
}

func (c *Client) Join(lobbyName string) {
	c.incomingPipe <- incomingMessage{
		what:    messageTypeJoin,
		client:  c,
		payload: lobbyName,
	}
}

func (c *Client) Part() {
	c.incomingPipe <- incomingMessage{
		what:   messageTypePart,
		client: c,
	}
}

func (c *Client) Ready() {
	c.incomingPipe <- incomingMessage{
		what:   messageTypeReady,
		client: c,
	}
}

func (c *Client) Word(word string) {
	c.incomingPipe <- incomingMessage{
		what:    messageTypeWord,
		client:  c,
		payload: word,
	}
}
