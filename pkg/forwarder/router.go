package forwarder

import (
	"context"

	"github.com/thyyl/chatr/pkg/common"
)

type Router struct {
	grpcServer common.GrpcServer
}

func NewRouter(grpcServer common.GrpcServer) *Router {
	return &Router{
		grpcServer: grpcServer,
	}
}

func (r *Router) Run() {
	r.grpcServer.RegisterServices()
	r.grpcServer.Run()
}

func (r *Router) GracefulStop(ctx context.Context) error {
	return r.grpcServer.GracefulStop()
}
