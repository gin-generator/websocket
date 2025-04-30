package websocket

type Option interface {
	apply(*Context)
}

type optionFunc func(*Context)

func (f optionFunc) apply(client *Context) {
	f(client)
}
