package service

import (
	"context"
	"time"

	"github.com/jxskiss/errors"
	"github.com/jxskiss/gopkg/set"

	"github.com/jxskiss/nonamegw/pkg/connid"
	"github.com/jxskiss/nonamegw/pkg/model"
	"github.com/jxskiss/nonamegw/proto/brokersvc"
	"github.com/jxskiss/nonamegw/proto/data"
	"github.com/jxskiss/nonamegw/proto/messag"
	"github.com/jxskiss/nonamegw/proto/protocol"
)

// TODO: app_id/app_secret auth middleware

func NewService(signer Signer, connDao ConnectionDao, nats NatsService) *Service {
	return &Service{
		signer:  signer,
		connDao: connDao,
		nats:    nats,
	}
}

type Service struct {
	signer  Signer
	connDao ConnectionDao
	nats    NatsService
}

func (p *Service) Query(ctx context.Context, request *brokersvc.QueryRequest) (*brokersvc.QueryResponse, error) {
	appId := request.GetAuth().GetAppId()
	userIds := request.GetUserIds()
	deviceIds := request.GetDeviceIds()

	var userConnections map[int64][]*data.ConnectionInfo
	var deviceConnections map[int64][]*data.ConnectionInfo
	var err error
	if len(userIds) > 0 {
		userConnections, err = p.connDao.ListUserConnections(ctx, appId, userIds)
		if err != nil {
			return nil, errors.AddStack(err)
		}
	}
	if len(deviceIds) > 0 {
		deviceConnections, err = p.connDao.ListDeviceConnections(ctx, appId, deviceIds)
		if err != nil {
			return nil, errors.AddStack(err)
		}
	}

	resp := &brokersvc.QueryResponse{}
	if len(userConnections) > 0 {
		resp.UserConnections = make(map[int64]*protocol.ConnectionList, len(userConnections))
		for userId, conns := range userConnections {
			resp.UserConnections[userId] = model.ToProtocolConnectionList(conns)
		}
	}
	if len(deviceConnections) > 0 {
		resp.DeviceConnections = make(map[int64]*protocol.ConnectionList, len(deviceConnections))
		for deviceId, conns := range deviceConnections {
			resp.DeviceConnections[deviceId] = model.ToProtocolConnectionList(conns)
		}
	}
	return resp, nil
}

func (p *Service) Push(ctx context.Context, request *brokersvc.PushRequest) (*brokersvc.PushResponse, error) {
	appId := request.GetAuth().GetAppId()
	target := request.GetTarget()
	content := request.GetContent()
	groupConnIds, err := p.resolvePushTarget(ctx, appId, target)
	if err != nil {
		return nil, err
	}
	packet := &protocol.Packet{
		BizFlag: content.GetBizFlag(),
		Headers: content.GetHeaderSlice(),
		Payload: content.GetPayload(),
	}
	for machineId, connIds := range groupConnIds {
		message := &messag.DowngoingMessage{
			Data: &messag.DowngoingMessage_Packet{
				Packet: packet,
			},
			ConnIds: connIds,
		}
		p.nats.PushGroupedMessage(machineId, message)
	}
	return &brokersvc.PushResponse{}, nil
}

func (p *Service) Sync(ctx context.Context, request *brokersvc.SyncRequest) (*brokersvc.SyncResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *Service) Broadcast(ctx context.Context, request *brokersvc.BroadcastRequest) (*brokersvc.BroadcastResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *Service) StopBroadcast(ctx context.Context, request *brokersvc.StopBroadcastRequest) (*brokersvc.StopBroadcastResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *Service) SignToken(ctx context.Context, request *brokersvc.SignTokenRequest) (*brokersvc.SignTokenResponse, error) {
	appId := request.GetAuth().GetAppId()
	userId := request.GetUserId()
	deviceId := request.GetDeviceId()

	nowTime := time.Now()
	token, err := p.signer.SignAuthToken(ctx, appId, userId, deviceId)
	if err != nil {
		return nil, errors.AddStack(err)
	}
	resp := &brokersvc.SignTokenResponse{
		Token:    token.Token,
		ExpireAt: nowTime.Add(TokenExpiration).Unix(),
	}
	return resp, nil
}

func (p *Service) resolvePushTarget(ctx context.Context, appId int64, target *brokersvc.PushTarget) (
	map[string][]string, error,
) {
	var connectionIds []string
	switch target.GetType() {
	case brokersvc.PushTarget_CONNECTION:
		connectionIds = target.GetConnections().GetConnectionIds()
	case brokersvc.PushTarget_USER:
		userConnections, err := p.connDao.ListUserConnections(ctx, appId, target.GetUsers().GetUserIds())
		if err != nil {
			return nil, err
		}
		connectionIds = make([]string, 0, len(userConnections))
		for _, conns := range userConnections {
			for _, c := range conns {
				connectionIds = append(connectionIds, c.Id)
			}
		}
	case brokersvc.PushTarget_USER_DEVICE:
		userIds := set.NewInt64()
		userDeviceIds := make(map[userDeviceId]struct{}, len(target.GetUserDevices().GetUserDevices()))
		for _, ud := range target.GetUserDevices().GetUserDevices() {
			userIds.Add(ud.UserId)
			udId := userDeviceId{userId: ud.UserId, deviceId: ud.DeviceId}
			userDeviceIds[udId] = struct{}{}
		}
		userConnections, err := p.connDao.ListUserConnections(ctx, appId, userIds.Slice())
		if err != nil {
			return nil, err
		}
		connectionIds = make([]string, 0, len(userConnections))
		for _, conns := range userConnections {
			for _, c := range conns {
				udId := userDeviceId{userId: c.UserId, deviceId: c.DeviceId}
				if _, ok := userDeviceIds[udId]; ok {
					connectionIds = append(connectionIds, c.Id)
				}
			}
		}
	case brokersvc.PushTarget_UNAUTHENTICATED_DEVICE:
		deviceConnections, err := p.connDao.ListDeviceConnections(ctx, appId, target.GetDevices().GetDeviceIds())
		if err != nil {
			return nil, err
		}
		connectionIds = make([]string, 0, len(deviceConnections))
		for _, conns := range deviceConnections {
			for _, c := range conns {
				connectionIds = append(connectionIds, c.Id)
			}
		}
	default:
		return nil, errors.Errorf("unknown target type %v", target.GetType())
	}
	if len(connectionIds) == 0 {
		return nil, nil
	}
	result := groupConnectionIds(connectionIds)
	return result, nil
}

func groupConnectionIds(connectionIds []string) map[string][]string {
	out := make(map[string][]string)
	for _, id := range connectionIds {
		connId, err := connid.ParseConnectionId(id)
		if err != nil {
			// TODO: logging
			continue
		}
		out[connId.MachineId] = append(out[connId.MachineId], id)
	}
	return out
}

type userDeviceId struct {
	userId, deviceId int64
}
