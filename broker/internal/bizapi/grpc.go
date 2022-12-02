package bizapi

import (
	"context"
	"github.com/jxskiss/errors"
	"github.com/jxskiss/nonamegw/broker/service"
	"github.com/jxskiss/nonamegw/proto/bizapi"
	"github.com/jxskiss/nonamegw/proto/protocol"
	"google.golang.org/grpc"
	"sync/atomic"
)

func NewBizApiImpl() service.BizApi {
	return &bizApiImpl{}
}

// FIXME
type bizApiImpl struct {
}

const (
	exampleChatAppId = 1001
	exampleChatAddr  = "127.0.0.1:9433"
)

var chatClient atomic.Value

func getChatClient() (bizapi.BizApiClient, error) {
	if client := chatClient.Load(); client != nil {
		return client.(bizapi.BizApiClient), nil
	}
	cc, err := grpc.Dial(exampleChatAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := bizapi.NewBizApiClient(cc)
	chatClient.Store(client)
	return client, nil
}

func (p *bizApiImpl) OnMessage(ctx context.Context, message *protocol.Message) error {
	conn := message.GetConn()
	if conn.GetAppId() != exampleChatAppId {
		return errors.Errorf("unknown app_id %v", conn.GetAppId())
	}
	bizCli, err := getChatClient()
	if err != nil {
		return errors.AddStack(err)
	}
	bizReq := &bizapi.OnMessageRequest{
		Message: message,
	}
	_, err = bizCli.OnMessage(ctx, bizReq)
	if err != nil {
		return errors.AddStack(err)
	}
	return nil
}

func (p *bizApiImpl) OnEvent(ctx context.Context, event *protocol.Event) error {
	conn := event.GetConn()
	if conn.GetAppId() != exampleChatAppId {
		return errors.Errorf("unknown app_id %v", conn.GetAppId())
	}
	bizCli, err := getChatClient()
	if err != nil {
		return errors.AddStack(err)
	}
	bizReq := &bizapi.OnEventRequest{
		Event: event,
	}
	_, err = bizCli.OnEvent(ctx, bizReq)
	if err != nil {
		return errors.AddStack(err)
	}
	return nil
}
