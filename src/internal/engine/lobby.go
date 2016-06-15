package engine

func (e *Engine) NewLobby(name, password string) *lobby {
	return &lobby{
		name:     name,
		password: password,
	}
}
