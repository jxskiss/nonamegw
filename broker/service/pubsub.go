package service

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/jxskiss/errors"
	"github.com/jxskiss/gopkg/easy"
	"github.com/nats-io/nats.go"

	"github.com/jxskiss/nonamegw/pkg/connid"
	"github.com/jxskiss/nonamegw/pkg/constants"
	"github.com/jxskiss/nonamegw/proto/cometsvc"
	"github.com/jxskiss/nonamegw/proto/messag"
	"github.com/jxskiss/nonamegw/proto/protocol"
)

// TODO: context

type NatsService interface {
	Close() error
	PushGroupedMessage(machineId string, message *messag.DowngoingMessage)
	PushMessages(messages []*messag.DowngoingMessage)
}

func NewNatsService(client *nats.Conn, bizapi BizApi, connDao ConnectionDao, signer Signer) (NatsService, error) {
	ec, err := nats.NewEncodedConn(client, "pb")
	if err != nil {
		return nil, err
	}
	impl := &natsImpl{
		client:  ec,
		bizapi:  bizapi,
		connDao: connDao,
		signer:  signer,
	}
	if err = impl.Setup(); err != nil {
		return nil, err
	}
	return impl, nil
}

type natsImpl struct {
	client  *nats.EncodedConn
	bizapi  BizApi
	connDao ConnectionDao
	signer  Signer
}

func (n *natsImpl) Setup() error {
	// rpc through Nats
	_, err := n.client.QueueSubscribe(
		constants.BrokerRpcWildcardTopic,
		constants.BrokerGroup,
		n.serveRPC)
	if err != nil {
		return errors.AddStack(err)
	}

	// consumer
	_, err = n.client.QueueSubscribe(constants.UpgoingMessageTopic, constants.BrokerGroup, n.handleMessage)
	if err != nil {
		return errors.AddStack(err)
	}
	_, err = n.client.QueueSubscribe(constants.EventTopic, constants.BrokerGroup, n.handleEvent)
	if err != nil {
		return errors.AddStack(err)
	}

	return nil
}

func (n *natsImpl) Close() error {
	// TODO
	return n.client.Drain()
}

func (n *natsImpl) serveRPC(subject, reply string, req *cometsvc.BrokerRequest) {
	switch subject {
	case constants.BrokerRpcGetCometConfigurationTopic:
		n.rpcGetCometConfiguration(reply, req.GetGetCometConfigurationRequest())
	case constants.BrokerRpcVerifyAuthTokenTopic:
		n.rpcVerifyAuthToken(reply, req.GetVerifyAuthTokenRequest())
	default:
		// TODO
	}
}

func (n *natsImpl) rpcGetCometConfiguration(reply string, req *cometsvc.GetCometConfigurationRequest) {
	// TODO
	_ = req
	resp := &cometsvc.GetCometConfigurationResponse{}
	err := n.client.Publish(reply, resp)

	// FIXME
	easy.PanicOnError(err)
}

func (n *natsImpl) rpcVerifyAuthToken(reply string, req *cometsvc.VerifyAuthTokenRequest) {
	resp := &cometsvc.VerifyAuthTokenResponse{}
	ctx := context.TODO()
	token, err := n.signer.DecodeAuthToken(ctx, req.GetToken())
	if err != nil {
		resp.Code = cometsvc.AuthToken_INVALID // TODO
	} else {
		resp.Code = cometsvc.AuthToken_SUCCESS
		resp.Token = token
	}
	err = n.client.Publish(reply, resp)

	// FIXME
	easy.PanicOnError(err)
}

func (n *natsImpl) handleMessage(msg *messag.UpgoingMessage) {
	ctx := context.TODO()
	packet := msg.GetPacket()
	bizMsg := &protocol.Message{
		Conn: msg.GetConn(),
		Content: &protocol.Content{
			BizFlag: packet.GetBizFlag(),
			Headers: packet.GetHeaderMap(),
			Payload: packet.GetPayload(),
		},
	}
	err := n.bizapi.OnMessage(ctx, bizMsg)
	if err != nil {
		// TODO
	}
}

func (n *natsImpl) handleEvent(event *protocol.Event) {
	ctx := context.TODO()
	err := n.bizapi.OnEvent(ctx, event)

	// TODO: manage connection data

	// FIXME
	easy.PanicOnError(err)
}

func (n *natsImpl) PushGroupedMessage(machineId string, message *messag.DowngoingMessage) {
	topic := constants.CometDowngoingMessageTopic(machineId)
	err := n.client.Publish(topic, message)
	if err != nil {
		// TODO
	}
}

func (n *natsImpl) PushMessages(messages []*messag.DowngoingMessage) {
	for _, msg := range messages {
		for _, id := range msg.ConnIds {
			connId, err := connid.ParseConnectionId(id)
			if err != nil {
				// TODO
			}
			machineId := connId.MachineId
			n.PushGroupedMessage(machineId, msg)
		}
	}
}

func init() {
	nats.RegisterEncoder("pb", ProtobufEncoder{})
}

type ProtobufEncoder struct{}

func (pe ProtobufEncoder) Encode(subject string, msg interface{}) ([]byte, error) {
	return proto.Marshal(msg.(proto.Message))
}

func (pe ProtobufEncoder) Decode(subject string, data []byte, msg interface{}) (err error) {
	return proto.Unmarshal(data, msg.(proto.Message))
}
