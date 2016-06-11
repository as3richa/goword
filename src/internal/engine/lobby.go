package engine

func (l *lobby) Send(m wrappedMessage) {
	l.pipe <- m
}
