package engine

func (c *Client) wrapMessage(m Message) wrappedMessage {
	return wrappedMessage{
		Client:  c,
		Message: m,
	}
}

func (e *Engine) NewClient() *Client {
	client := &Client{
		container: e,
		Pipe:      make(chan Response, clientBufferSize),
	}

	e.Send(client.wrapMessage(connectMessage{}))

	return client
}

func (c *Client) Send(r Response) {
	c.Pipe <- r
}
