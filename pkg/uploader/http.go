package uploader

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/config"
)

type HttpServer struct {
	name                     string
	logger                   common.HttpLog
	server                   *gin.Engine
	httpServer               *http.Server
	httpPort                 string
	s3Endpoint               string
	s3Bucket                 string
	maxMemory                int64
	uploader                 *manager.Uploader
	presigner                *Presigner
	channelUploadRateLimiter ChannelUploadRateLimiter
	serveSwag                bool
}

func NewGinServer(name string, logger common.HttpLog, config *config.Config) *gin.Engine {
	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(common.CorsMiddleware())
	server.Use(common.LoggingMiddleware(logger))

	return server
}

func NewHttpServer(name string, logger common.HttpLog, config *config.Config, server *gin.Engine, channelUploadRateLimiter ChannelUploadRateLimiter) *HttpServer {
	s3Endpoint := config.Uploader.S3.Endpoint
	s3Bucket := config.Uploader.S3.Bucket
	credentials := credentials.NewStaticCredentialsProvider(config.Uploader.S3.AccessKey, config.Uploader.S3.SecretKey, "")
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               s3Endpoint,
			SigningRegion:     config.Uploader.S3.Region,
			HostnameImmutable: true,
		}, nil
	})
	awsConfig := aws.Config{
		Credentials:                 credentials,
		EndpointResolverWithOptions: customResolver,
		Region:                      config.Uploader.S3.Region,
		RetryMaxAttempts:            3,
	}
	s3Client := s3.NewFromConfig(awsConfig)

	return &HttpServer{
		name:                     name,
		logger:                   logger,
		server:                   server,
		s3Endpoint:               s3Endpoint,
		s3Bucket:                 s3Bucket,
		maxMemory:                config.Uploader.Http.Server.MaxMemoryByte,
		uploader:                 manager.NewUploader(s3Client),
		presigner:                &Presigner{s3.NewPresignClient(s3Client), config.Uploader.S3.PresignLifetimeSecond},
		httpPort:                 config.Uploader.Http.Server.Port,
		channelUploadRateLimiter: channelUploadRateLimiter,
		serveSwag:                config.Uploader.Http.Server.Swag,
	}
}

func (s *HttpServer) ChannelUploadRateLimit() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channelID, ok := ctx.Request.Context().Value(common.ChannelKey).(uint64)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		allow, err := s.channelUploadRateLimiter.Allow(ctx.Request.Context(), strconv.FormatUint(channelID, 10))

		if err != nil {
			s.logger.Error(err.Error())
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if !allow {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}

func (s *HttpServer) RegisterRoutes() {
	uploaderGroup := s.server.Group("/api/uploader")
	{
		uploadGroup := uploaderGroup.Group("/upload")
		uploadGroup.Use(common.JWTForwardAuth())
		uploadGroup.Use(s.ChannelUploadRateLimit())
		{
			uploadGroup.POST("/files", s.UploadFiles)
			uploadGroup.GET("/presigned", s.GetPresignedUpload)
		}

		downloadGroup := uploaderGroup.Group("/download")
		downloadGroup.Use(common.JWTForwardAuth())
		{
			downloadGroup.GET("/presigned", s.GetPresignedDownload)
		}
	}
}

func (s *HttpServer) Run() {
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
}
func (r *HttpServer) GracefulStop(ctx context.Context) error {
	return r.httpServer.Shutdown(ctx)
}
