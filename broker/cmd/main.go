package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/jxskiss/nonamegw/broker/service"
	"github.com/jxskiss/nonamegw/pkg/zlog"
	"github.com/jxskiss/nonamegw/proto/brokersvc"
)

func main() {
	logger, prop, _ := zlog.NewLogger(&zlog.Config{
		Level:       "debug",
		Format:      "console",
		Development: true,
	})
	zlog.ReplaceGlobals(logger, prop)

	app, err := InitApp()
	if err != nil {
		zlog.Fatalf("failed init application, err= %v", err)
	}
	rpcServer := grpc.NewServer()
	brokersvc.RegisterBrokerServer(rpcServer, app.rpcImpl)
	zlog.Infof("starting broker/rpc server listening on %v", cfg.RpcListen)
	go func() {
		ln, err := net.Listen("tcp", cfg.RpcListen)
		if err != nil {
			zlog.Fatalf("failed listen broker/rpc, err= %v", err)
		}
		err = rpcServer.Serve(ln)
		if err != nil {
			zlog.Fatalf("failed serving broker/rpc, err= %v", err)
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	<-exit
	rpcServer.GracefulStop()
}

// ---- configuration ---- //

var cfg = &Config{
	RpcListen: "127.0.0.1:9432",
}

type Config struct {
	RpcListen string
}

// ---- application ---- //

func NewApp(nats service.NatsService, rpcImpl brokersvc.BrokerServer) *App {
	return &App{
		nats:    nats,
		rpcImpl: rpcImpl,
	}
}

type App struct {
	nats    service.NatsService
	rpcImpl brokersvc.BrokerServer
}
