package engine

const clientBufferSize = 16

type Client struct {
	alive        bool
	lobby        *lobby
	messagePipe  chan wrappedMessage
	ResponsePipe chan Response
}

func (c *Client) wrapMessage(m Message) wrappedMessage {
	return wrappedMessage{
		Client:  c,
		Message: m,
	}
}

func (e *Engine) NewClient() *Client {
	client := &Client{
		alive:        true,
		messagePipe:  e.pipe,
		ResponsePipe: make(chan Response, clientBufferSize),
	}

	client.SendFrom(connectMessage{})

	return client
}

func (c *Client) SendTo(r Response) {
	c.ResponsePipe <- r
}

func (c *Client) SendFrom(m Message) {
	c.messagePipe <- c.wrapMessage(m)
}

func (c *Client) Quit() {
	c.SendFrom(quitMessage{})
}
