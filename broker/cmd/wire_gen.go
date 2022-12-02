// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package main

import (
	"github.com/jxskiss/nonamegw/broker/adapter"
	"github.com/jxskiss/nonamegw/broker/internal/bizapi"
	"github.com/jxskiss/nonamegw/broker/internal/dao"
	"github.com/jxskiss/nonamegw/broker/internal/infra"
	"github.com/jxskiss/nonamegw/broker/service"
)

// Injectors from wire.go:

func InitApp() (*App, error) {
	conn, err := infra.InitNatsClient()
	if err != nil {
		return nil, err
	}
	bizApi := bizapi.NewBizApiImpl()
	client, err := infra.InitRedis()
	if err != nil {
		return nil, err
	}
	connectionDao := dao.NewConnectionDao(client)
	tokenDao := dao.NewTokenDao(client)
	signer := service.NewSigner(tokenDao)
	natsService, err := service.NewNatsService(conn, bizApi, connectionDao, signer)
	if err != nil {
		return nil, err
	}
	serviceService := service.NewService(signer, connectionDao, natsService)
	brokerServer := adapter.NewRpcImpl(serviceService)
	app := NewApp(natsService, brokerServer)
	return app, nil
}