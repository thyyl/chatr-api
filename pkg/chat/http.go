package chat

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

var MelodyChat MelodyChatConn

type MelodyChatConn struct {
	*melody.Melody
}

func NewMelodyChat(config *config.Config) MelodyChatConn {
	melody := melody.New()
	melody.Config.MaxMessageSize = config.Chat.Message.MaxSizeByte
	MelodyChat = MelodyChatConn{
		melody,
	}

	return MelodyChat
}

type HttpServer struct {
	name              string
	logger            common.HttpLog
	server            *gin.Engine
	melodyChat        MelodyChatConn
	httpPort          string
	httpServer        *http.Server
	messageSubscriber *MessageSubscriber
	userService       UserService
	chatService       ChatService
	channelService    ChannelService
	forwarderService  ForwarderService
	serveSwag         bool
}

func NewGinServer(name string, logger common.HttpLog, config *config.Config) *gin.Engine {
	server := gin.New()
	server.Use(gin.Logger())
	server.Use(gin.Recovery())
	server.Use(common.CorsMiddleware())
	server.Use(common.MaxAllowed(config.Chat.Http.Server.MaxConn))

	return server
}

func NewHttpServer(name string, logger common.HttpLog, config *config.Config, server *gin.Engine, melody MelodyChatConn, messageSubscriber *MessageSubscriber, userService UserService, chatService ChatService, channelService ChannelService, forwarderService ForwarderService) *HttpServer {
	common.JwtSecret = config.Chat.JWT.Secret
	common.JwtExpirationSecond = config.Chat.JWT.ExpirationSecond

	return &HttpServer{
		name:              name,
		logger:            logger,
		server:            server,
		melodyChat:        melody,
		httpPort:          config.Chat.Http.Server.Port,
		messageSubscriber: messageSubscriber,
		userService:       userService,
		chatService:       chatService,
		channelService:    channelService,
		forwarderService:  forwarderService,
		serveSwag:         config.Chat.Http.Server.Swag,
	}
}

func (s *HttpServer) RegisterRoutes() {
	s.messageSubscriber.RegisterHandler()

	chatGroup := s.server.Group("/api/chat")
	{
		chatGroup.GET("", s.StartChat)

		forwarderAuthGroup := chatGroup.Group("/forwarderauth")
		forwarderAuthGroup.Use(common.JWTAuth())
		{
			forwarderAuthGroup.GET("", s.ForwardAuth)
		}

		userGroup := chatGroup.Group("/user")
		userGroup.Use(common.JWTAuth())
		{
			userGroup.GET("", s.GetChannelUsers)
			userGroup.GET("/online", s.GetOnlineUsers)
		}

		channelGroup := chatGroup.Group("/channel")
		channelGroup.Use(common.JWTAuth())
		{
			channelGroup.GET("/messages", s.ListMessages)
			channelGroup.DELETE("", s.DeleteChannel)
		}
	}

	s.melodyChat.HandleConnect(s.HandleChatOnConnect)
	s.melodyChat.HandleMessage(s.HandleChatOnMessage)
	s.melodyChat.HandleClose(s.HandleChatOnClose)
}

func (s *HttpServer) Run() {
	go func() {
		addr := ":" + s.httpPort
		s.httpServer = &http.Server{
			Addr:    addr,
			Handler: s.server,
		}
		s.logger.Info("http server listening", slog.String("addr", addr))
		err := s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.logger.Error(err.Error())
			os.Exit(1)
		}
	}()
	go func() {
		err := s.messageSubscriber.Run()
		if err != nil {
			s.logger.Error(err.Error())
			os.Exit(1)
		}
	}()
}

func (s *HttpServer) GracefulStop(ctx context.Context) error {
	err := MelodyChat.Close()
	if err != nil {
		return err
	}
	err = s.httpServer.Shutdown(ctx)
	if err != nil {
		return err
	}
	err = s.messageSubscriber.GracefulStop()
	if err != nil {
		return err
	}
	return nil
}
