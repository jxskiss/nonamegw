package service

import (
	"context"
	"github.com/jxskiss/nonamegw/proto/data"
)

/*
路由索引

(app_id, user_id)
用于推送登录用户及登录用户的指定设备。
hash
- Key: h:u:{app_id}:{user_id}
- Hash key: connection_id
- Hash Value: connection meta information
- HSET KEY connection_id meta 写入索引
- HGETALL KEY 查询一个用户的所有连接
- HDEL KEY connection_id 删除连接
zset
- Key: s:u:{app_id}:{user_id}
- Member: connection_id
- Score: connection last update time

(app_id, device_id)
用于推送非登录用户。
hash
- Key: "h:d:{app_id}:{device_id}"
- Hash key/value 格式和操作同 (app_id, user_id) 索引
zset
- Key: "s:d:{app_id}:{device_id}"
- Member, Score 同 (app_id, user_id) 索引

过期连接清理
每次有连接活跃时，清理不活跃连接
- zrevrange zsetKey 0 5 (以最多保留5个连接为例)
- IF len(members) > 5:
  - Get outdated connection IDs from zrevrange zsetKey 5 -1 (仅保留5个连接)
  - zremrangebyrank zsetKey 5 -1
  - hdel hashKey [outdated connection IDs]...
- Always do:
  - expire hashKey purgeTTL
  - expire zsetKey purgeTTL

临时连接
临时连接不关联 user_id / device_id，不接受 user_id / device_id 推送及广播推送。
不需要在 Redis 中维护路由表，使用 Connection ID 直接推送即可。
*/

type ConnectionDao interface {
	SaveConnection(ctx context.Context, connection *data.ConnectionInfo) error
	DeleteConnection(ctx context.Context, appId, userId, deviceId int64, connectionId string) error
	TouchConnection(ctx context.Context, appId, userId, deviceId int64, connectionId string) error

	ListUserConnections(ctx context.Context, appId int64, userIds []int64) (map[int64][]*data.ConnectionInfo, error)
	ListDeviceConnections(ctx context.Context, appId int64, deviceIds []int64) (map[int64][]*data.ConnectionInfo, error)
}
