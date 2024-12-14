package user

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type HttpServer struct {
	name              string
	logger            common.HttpLog
	server            *gin.Engine
	httpServer        *http.Server
	httpPort          string
	serveSwag         bool
	userService       UserService
	googleOAuthConfig *oauth2.Config
	oAuthCookieConfig config.CookieConfig
	authCookieConfig  config.CookieConfig
}

func NewGinServer(name string, logger common.HttpLog, config *config.Config) *gin.Engine {
	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(common.CorsMiddleware())
	server.Use(common.LoggingMiddleware(logger))

	return server
}

func NewHttpServer(name string, logger common.HttpLog, config *config.Config, server *gin.Engine, userService UserService) *HttpServer {
	return &HttpServer{
		name:        name,
		logger:      logger,
		server:      server,
		httpPort:    config.Users.Http.Server.Port,
		serveSwag:   config.Users.Http.Server.Swag,
		userService: userService,
		googleOAuthConfig: &oauth2.Config{
			RedirectURL:  config.Users.OAuth.Google.RedirectUrl,
			ClientID:     config.Users.OAuth.Google.ClientId,
			ClientSecret: config.Users.OAuth.Google.ClientSecret,
			Scopes:       strings.Split(config.Users.OAuth.Google.Scopes, ","),
			Endpoint:     google.Endpoint,
		},
		oAuthCookieConfig: config.Users.OAuth.Cookie,
		authCookieConfig:  config.Users.Auth.Cookie,
	}
}

func (s *HttpServer) RegisterRoutes() {
	userGroup := s.server.Group("/api/user")
	{
		userGroup.POST("/", s.CreateLocalUser)

		userGroup.GET("/oauth2/google/login", s.OAuthGoogleLogin)
		userGroup.GET("/oauth2/google/callback", s.OAuthGoogleCallback)

		authGroup := userGroup.Group("")
		authGroup.Use(s.CookieAuth())
		authGroup.GET("", s.GetUser)
		authGroup.GET("/me", s.GetUserMe)
	}
}

func (s *HttpServer) Run() {
	go func() {
		address := ":" + s.httpPort
		s.httpServer = &http.Server{
			Addr:    address,
			Handler: s.server,
		}

		s.logger.Info("Starting HTTP server", "address", address)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error(err.Error())
			os.Exit(1)
		}
	}()
}

func (s *HttpServer) GracefulStop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *HttpServer) CookieAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session, err := common.GetCookie(ctx, common.SessionIdCookieName)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userId, err := s.userService.GetUserIdBySession(ctx.Request.Context(), session)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), common.UserKey, userId))
		ctx.Next()
	}
}
