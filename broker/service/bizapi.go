package service

import (
	"context"
	"github.com/jxskiss/nonamegw/proto/protocol"
)

type BizApi interface {
	OnMessage(ctx context.Context, message *protocol.Message) error
	OnEvent(ctx context.Context, event *protocol.Event) error
}
