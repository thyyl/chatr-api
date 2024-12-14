package common

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HttpServer interface {
	RegisterRoutes()
	Run()
	GracefulStop(ctx context.Context) error
}

type GrpcServer interface {
	RegisterServices()
	Run()
	GracefulStop() error
}

type Router interface {
	Run()
	GracefulStop(ctx context.Context) error
}

type InfraCloser interface {
	Close() error
}

type Server struct {
	name        string
	router      Router
	infraCloser InfraCloser
}

func NewServer(name string, router Router, infraCloser InfraCloser) *Server {
	return &Server{
		name:        name,
		router:      router,
		infraCloser: infraCloser,
	}
}

func (s *Server) Serve() {
	s.router.Run()
	slog.Info("Server started")

	done := make(chan bool, 1)

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.GracefulStop(ctx, done)
	}()

	<-done
}

func (s *Server) GracefulStop(ctx context.Context, done chan bool) {
	err := s.router.GracefulStop(ctx)

	if err != nil {
		slog.Error(err.Error())
	}

	if err = s.infraCloser.Close(); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Server stopped gracefully")
	done <- true
}
