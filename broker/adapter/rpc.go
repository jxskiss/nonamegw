package adapter

import (
	"context"
	"github.com/jxskiss/nonamegw/broker/service"
	"github.com/jxskiss/nonamegw/proto/brokersvc"
)

var _ brokersvc.BrokerServer = &RpcImpl{}

func NewRpcImpl(svc *service.Service) brokersvc.BrokerServer {
	return &RpcImpl{svc: svc}
}

type RpcImpl struct {
	brokersvc.UnimplementedBrokerServer

	svc *service.Service
}

func (r *RpcImpl) Query(ctx context.Context, request *brokersvc.QueryRequest) (*brokersvc.QueryResponse, error) {
	return r.svc.Query(ctx, request)
}

func (r *RpcImpl) Push(ctx context.Context, request *brokersvc.PushRequest) (*brokersvc.PushResponse, error) {
	return r.svc.Push(ctx, request)
}

func (r *RpcImpl) Sync(ctx context.Context, request *brokersvc.SyncRequest) (*brokersvc.SyncResponse, error) {
	return r.svc.Sync(ctx, request)
}

func (r *RpcImpl) Broadcast(ctx context.Context, request *brokersvc.BroadcastRequest) (*brokersvc.BroadcastResponse, error) {
	return r.svc.Broadcast(ctx, request)
}

func (r *RpcImpl) StopBroadcast(ctx context.Context, request *brokersvc.StopBroadcastRequest) (*brokersvc.StopBroadcastResponse, error) {
	return r.svc.StopBroadcast(ctx, request)
}

func (r *RpcImpl) SignToken(ctx context.Context, request *brokersvc.SignTokenRequest) (*brokersvc.SignTokenResponse, error) {
	return r.svc.SignToken(ctx, request)
}
