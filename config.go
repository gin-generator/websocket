package websocket

const (
	RegisterLimit = 100
)

type Config struct {
	Host string
	Port string
}

type ManagerCfg struct {
	RegisterLimit int
}
