//+build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/jxskiss/nonamegw/broker/adapter"
	"github.com/jxskiss/nonamegw/broker/internal/bizapi"
	"github.com/jxskiss/nonamegw/broker/internal/dao"
	"github.com/jxskiss/nonamegw/broker/internal/infra"
	"github.com/jxskiss/nonamegw/broker/service"
)

func InitApp() (*App, error) {
	wire.Build(
		NewApp,
		adapter.NewRpcImpl,
		service.NewNatsService,
		service.NewService,
		service.NewSigner,
		dao.NewTokenDao,
		dao.NewConnectionDao,
		bizapi.NewBizApiImpl,
		infra.InitNatsClient,
		infra.InitRedis,
	)
	return &App{}, nil
}
