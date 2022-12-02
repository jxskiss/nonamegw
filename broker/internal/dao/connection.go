package dao

import (
	"context"
	stderr "errors"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	"github.com/jxskiss/errors"

	"github.com/jxskiss/nonamegw/broker/service"
	"github.com/jxskiss/nonamegw/proto/data"
)

const (
	purgeExpiration = 10 * time.Minute
)

var (
	ErrInvalidUserIdDeviceId = stderr.New("invalid user_id/device_id")
)

func NewConnectionDao(redisClient *redis.Client) service.ConnectionDao {
	return &connectionDaoImpl{
		redisCli: redisClient,
	}
}

type connectionDaoImpl struct {
	redisCli *redis.Client
}

func (p *connectionDaoImpl) SaveConnection(ctx context.Context, connection *data.ConnectionInfo) error {
	if connection.UserId <= 0 && connection.DeviceId <= 0 {
		return errors.AddStack(ErrInvalidUserIdDeviceId)
	}
	buf, err := proto.Marshal(connection)
	if err != nil {
		return errors.AddStack(err)
	}

	pipe := p.redisCli.Pipeline()
	defer pipe.Close()

	hkey, zkey := p.getConnectionKeys(connection.AppId, connection.UserId, connection.DeviceId)
	pipe.HSet(ctx, hkey, connection.Id, buf)
	pipe.ZAdd(ctx, zkey, &redis.Z{
		Score:  getTimeNowScore(),
		Member: connection.Id,
	})
	pipe.Expire(ctx, hkey, purgeExpiration)
	pipe.Expire(ctx, zkey, purgeExpiration)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return errors.AddStack(err)
	}

	go func() {
		p.purgeOutdatedConnections(ctx, hkey, zkey)
	}()

	return nil
}

func (p *connectionDaoImpl) purgeOutdatedConnections(ctx context.Context, hkey, zkey string) {
	invalidMembers, err := p.redisCli.ZRevRangeWithScores(ctx, zkey, 5, -1).Result()
	if err != nil {
		// TODO: logging
		return
	}
	if len(invalidMembers) == 0 {
		return
	}
	purgeScore := invalidMembers[0].Score
	invalidConnIds := make([]string, 0, len(invalidMembers))
	for _, m := range invalidMembers {
		invalidConnIds = append(invalidConnIds, m.Member.(string))
	}
	_, err = p.redisCli.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		maxScore := strconv.FormatFloat(purgeScore, 'f', 6, 64)
		pipe.HDel(ctx, hkey, invalidConnIds...)
		pipe.ZRemRangeByScore(ctx, zkey, "0", maxScore)
		return nil
	})
	if err != nil {
		// TODO: logging
	}
}

func (p *connectionDaoImpl) getConnectionKeys(appId, userId, deviceId int64) (hkey, zkey string) {
	if userId > 0 {
		hkey = userConnectionsHashKey(appId, userId)
		zkey = userConnectionsZsetKey(appId, userId)
		return
	}
	hkey = deviceConnectionsHashKey(appId, deviceId)
	zkey = deviceConnectionsZsetKey(appId, deviceId)
	return
}

func (p *connectionDaoImpl) DeleteConnection(ctx context.Context, appId, userId, deviceId int64, connectionId string) error {
	if userId <= 0 && deviceId <= 0 {
		return errors.AddStack(ErrInvalidUserIdDeviceId)
	}
	hkey, zkey := p.getConnectionKeys(appId, userId, deviceId)
	_, err := p.redisCli.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HDel(ctx, hkey, connectionId)
		pipe.ZRem(ctx, zkey, connectionId)
		return nil
	})
	if err != nil {
		return errors.AddStack(err)
	}
	return nil
}

func (p *connectionDaoImpl) TouchConnection(ctx context.Context, appId, userId, deviceId int64, connectionId string) error {
	if userId <= 0 && deviceId <= 0 {
		return errors.AddStack(ErrInvalidUserIdDeviceId)
	}
	hkey, zkey := p.getConnectionKeys(appId, userId, deviceId)
	_, err := p.redisCli.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.ZAdd(ctx, zkey, &redis.Z{
			Score:  getTimeNowScore(),
			Member: connectionId,
		})
		pipe.Expire(ctx, hkey, purgeExpiration)
		pipe.Expire(ctx, zkey, purgeExpiration)
		return nil
	})
	if err != nil {
		return errors.AddStack(err)
	}
	return nil
}

func (p *connectionDaoImpl) ListUserConnections(ctx context.Context, appId int64, userIds []int64) (map[int64][]*data.ConnectionInfo, error) {
	if len(userIds) == 0 {
		return nil, nil
	}

	hkeys := make([]string, 0, len(userIds))
	for _, userId := range userIds {
		hkeys = append(hkeys, userConnectionsHashKey(appId, userId))
	}

	pipe := p.redisCli.Pipeline()
	defer pipe.Close()

	for _, hk := range hkeys {
		pipe.HGetAll(ctx, hk)
	}
	pipeResult, err := pipe.Exec(ctx)
	if err != nil {
		return nil, errors.AddStack(err)
	}

	var resultConnections = make(map[int64][]*data.ConnectionInfo)
	for _, cmd := range pipeResult {
		val := cmd.(*redis.StringStringMapCmd).Val()
		for _, buf := range val {
			connInfo := &data.ConnectionInfo{}
			err := proto.Unmarshal([]byte(buf), connInfo)
			if err != nil {
				// TODO: logging
				continue
			}
			userId := connInfo.UserId
			resultConnections[userId] = append(resultConnections[userId], connInfo)
		}
	}
	return resultConnections, nil
}

func (p *connectionDaoImpl) ListDeviceConnections(ctx context.Context, appId int64, deviceIds []int64) (map[int64][]*data.ConnectionInfo, error) {
	if len(deviceIds) == 0 {
		return nil, nil
	}

	hkeys := make([]string, 0, len(deviceIds))
	for _, deviceId := range deviceIds {
		hkeys = append(hkeys, deviceConnectionsHashKey(appId, deviceId))
	}

	pipe := p.redisCli.Pipeline()
	defer pipe.Close()

	for _, hk := range hkeys {
		pipe.HGetAll(ctx, hk)
	}
	pipeResult, err := pipe.Exec(ctx)
	if err != nil {
		return nil, errors.AddStack(err)
	}

	var resultConnections = make(map[int64][]*data.ConnectionInfo)
	for _, cmd := range pipeResult {
		val := cmd.(*redis.StringStringMapCmd).Val()
		for _, buf := range val {
			connInfo := &data.ConnectionInfo{}
			err := proto.Unmarshal([]byte(buf), connInfo)
			if err != nil {
				// TODO: logging
				continue
			}
			deviceId := connInfo.DeviceId
			resultConnections[deviceId] = append(resultConnections[deviceId], connInfo)
		}
	}
	return resultConnections, nil
}

func getTimeNowScore() float64 {
	return float64(time.Now().Unix())
}
