package websocket

type Subscribe interface {
	Subscribe(channel string) error
	Unsubscribe(channel string) error
	Publish(channel string, message []byte) error
}

type Storage interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
}
