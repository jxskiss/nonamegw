package infra

import "github.com/nats-io/nats.go"

const natsServerAddr = "127.0.0.1:4222"

func InitNatsClient() (*nats.Conn, error) {
	return nats.Connect(natsServerAddr)
}
