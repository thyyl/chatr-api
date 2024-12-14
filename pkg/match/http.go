package match

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/config"
	"gopkg.in/olahol/melody.v1"
)

var (
	MelodyMatch MelodyMatchConn
)

type MelodyMatchConn struct {
	*melody.Melody
}

type HttpServer struct {
	name            string
	logger          common.HttpLog
	server          *gin.Engine
	melodyMatch     MelodyMatchConn
	httpPort        string
	httpServer      *http.Server
	matchSubscriber *MatchSubscriber
	userService     UserService
	matchService    MatchService
	serveSwag       bool
}

func NewMelodyMatchConn() MelodyMatchConn {
	MelodyMatch = MelodyMatchConn{
		melody.New(),
	}
	return MelodyMatch
}

func NewGinServer(name string, logger common.HttpLog, config *config.Config) *gin.Engine {
	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(common.CorsMiddleware())
	server.Use(common.LoggingMiddleware(logger))
	server.Use(common.MaxAllowed(config.Match.Http.Server.MaxConn))

	return server
}

func NewHttpServer(name string, logger common.HttpLog, config *config.Config, server *gin.Engine, melodyMatch MelodyMatchConn, matchSubscriber *MatchSubscriber, userService UserService, matchService MatchService) *HttpServer {
	return &HttpServer{
		name:            name,
		logger:          logger,
		server:          server,
		melodyMatch:     melodyMatch,
		httpPort:        config.Match.Http.Server.Port,
		matchSubscriber: matchSubscriber,
		userService:     userService,
		matchService:    matchService,
		serveSwag:       config.Match.Http.Server.Swag,
	}
}

func (s *HttpServer) CookieAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sid, err := common.GetCookie(c, common.SessionIdCookieName)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		userID, err := s.userService.GetUserIdBySession(c.Request.Context(), sid)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), common.UserKey, userID))
		c.Next()
	}
}

func (s *HttpServer) RegisterRoutes() {
	s.matchSubscriber.RegisterHandler()

	matchGroup := s.server.Group("/api/match")
	{
		authGroup := matchGroup.Group("")
		authGroup.Use(s.CookieAuth())
		authGroup.GET("", s.Match)
	}

	s.melodyMatch.HandleConnect(s.HandleMatchOnConnect)
	s.melodyMatch.HandleClose(s.HandleClose)
}

func (s *HttpServer) Run() {
	go func() {
		address := ":" + s.httpPort
		s.httpServer = &http.Server{
			Addr:    address,
			Handler: s.server,
		}

		s.logger.Info("http server is running on ", slog.String("address", address))

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error(err.Error())
			os.Exit(1)
		}
	}()

	go func() {
		err := s.matchSubscriber.Run()
		if err != nil {
			s.logger.Error(err.Error())
			os.Exit(1)
		}
	}()
}

func (s *HttpServer) GracefulStop(ctx context.Context) error {
	err := MelodyMatch.Close()
	if err != nil {
		return err
	}

	err = s.httpServer.Shutdown(ctx)
	if err != nil {
		return err
	}

	err = s.matchSubscriber.GracefulStop()
	if err != nil {
		return err
	}

	return nil
}
