package user

import (
	"context"

	"github.com/thyyl/chatr/pkg/common"
)

type Router struct {
	httpServer common.HttpServer
	grpcServer common.GrpcServer
}

func NewRouter(httpServer common.HttpServer, grpcServer common.GrpcServer) *Router {
	return &Router{httpServer, grpcServer}
}

func (router *Router) Run() {
	router.httpServer.RegisterRoutes()
	router.httpServer.Run()

	router.grpcServer.RegisterServices()
	router.grpcServer.Run()
}

func (router *Router) GracefulStop(ctx context.Context) error {
	if err := router.grpcServer.GracefulStop(); err != nil {
		return err
	}
	return router.httpServer.GracefulStop(ctx)
}
