package model

import (
	"github.com/jxskiss/nonamegw/proto/data"
	"github.com/jxskiss/nonamegw/proto/protocol"
)

func ToProtocolConnection(conn *data.ConnectionInfo) *protocol.Connection {
	return &protocol.Connection{
		Id:            conn.Id,
		AppId:         conn.AppId,
		UserId:        conn.UserId,
		DeviceId:      conn.DeviceId,
		ClientIp:      conn.ClientIp,
		ClientVersion: conn.ClientVersion,
	}
}

func ToProtocolConnectionList(conns []*data.ConnectionInfo) *protocol.ConnectionList {
	out := &protocol.ConnectionList{
		Connections: make([]*protocol.Connection, 0, len(conns)),
	}
	for _, c := range conns {
		out.Connections = append(out.Connections, ToProtocolConnection(c))
	}
	return out
}
