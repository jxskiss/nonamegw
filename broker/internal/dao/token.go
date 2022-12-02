package dao

import (
	"context"
	stderr "errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	"github.com/jxskiss/errors"

	"github.com/jxskiss/nonamegw/broker/service"
	"github.com/jxskiss/nonamegw/proto/data"
)

var ErrTokenNotExists = stderr.New("token not exists")

func NewTokenDao(redisClient *redis.Client) service.TokenDao {
	return &tokenDaoImpl{
		redisCli: redisClient,
	}
}

type tokenDaoImpl struct {
	redisCli *redis.Client
}

func (p *tokenDaoImpl) GetToken(ctx context.Context, token string) (*data.TokenInfo, error) {
	key := tokenKey(token)
	val, err := p.redisCli.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrTokenNotExists
		}
		return nil, errors.AddStack(err)
	}
	tokenInfo := &data.TokenInfo{}
	err = proto.Unmarshal(val, tokenInfo)
	if err != nil {
		return nil, errors.AddStack(err)
	}
	return tokenInfo, nil
}

func (p *tokenDaoImpl) SaveToken(ctx context.Context, info *data.TokenInfo, ttl time.Duration) error {
	key := tokenKey(info.Id)
	buf, err := proto.Marshal(info)
	if err != nil {
		return errors.AddStack(err)
	}
	err = p.redisCli.Set(ctx, key, buf, ttl).Err()
	if err != nil {
		return errors.AddStack(err)
	}
	return nil
}
