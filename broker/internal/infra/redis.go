package infra

import (
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
	"github.com/jxskiss/nonamegw/pkg/zlog"
	"time"
)

const miniRedisAddr = "127.0.0.1:6379"

func InitRedis() (*redis.Client, error) {
	zlog.Infof("running miniredis on %v", miniRedisAddr)

	s := miniredis.NewMiniRedis()
	err := s.StartAddr(miniRedisAddr)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr:         miniRedisAddr,
		DialTimeout:  50 * time.Millisecond,
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
		IdleTimeout:  time.Second,
	})
	return client, nil
}
