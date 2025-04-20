package websocket

type Option interface {
	apply(*Client)
}

type optionFunc func(*Client)

func (f optionFunc) apply(client *Client) {
	f(client)
}
