package adapter

import (
	"github.com/gin-gonic/gin"
	"github.com/jxskiss/nonamegw/broker/service"
)

type HttpServer struct {
	svc *service.Service
}

func (p *HttpServer) Query(c *gin.Context) {
	panicTodo()
}

func (p *HttpServer) Push(c *gin.Context) {
	panicTodo()
}

func (p *HttpServer) Sync(c *gin.Context) {
	panicTodo()
}

func (p *HttpServer) Broadcast(c *gin.Context) {
	panicTodo()
}

func (p *HttpServer) StopBroadcast(c *gin.Context) {
	panicTodo()
}

func (p *HttpServer) SignToken(c *gin.Context) {
	panicTodo()
}

func panicTodo() {
	panic("TODO: implementation")
}
