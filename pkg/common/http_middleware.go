package common

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type HTTPContextKey string

func MaxAllowed(n int64) gin.HandlerFunc {
	sem := make(chan struct{}, n)
	acquire := func() { sem <- struct{}{} }
	release := func() { <-sem }
	return func(c *gin.Context) {
		acquire()       // before request
		defer release() // after request
		c.Next()

	}
}

func CorsMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", JWTAuthHeader},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	return cors.New(config)
}

func LoggingMiddleware(logger HttpLog) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process Request
		c.Next()

		// Stop timer
		duration := getDurationInMillseconds(start)

		logger.Info("",
			slog.Float64("duration_ms", duration),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.RequestURI),
			slog.Int("status", c.Writer.Status()),
			slog.String("referrer", c.Request.Referer()),
			slog.String("trace_id", getTraceId(c)))
	}
}

func LimitBodySize(maxBodyBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
		c.Next()
	}
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := extractTokenFromHeader(c.Request)
		if accessToken == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		authResult, err := Auth(&AuthPayload{
			AccessToken: accessToken,
		})
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if authResult.Expired {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Message: ErrorTokenExpired.Error(),
			})
			return
		}
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ChannelKey, authResult.ChannelId))
		c.Next()
	}
}

func JWTForwardAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		channelId, err := strconv.ParseUint(c.Request.Header.Get(ChannelIdHeader), 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ChannelKey, channelId))
		c.Next()
	}
}

func getTraceId(c *gin.Context) string {
	identifier := c.Request.Header.Get(JaegerHeader)
	vals := strings.Split(identifier, ":")
	if len(vals) == 4 {
		return vals[0]
	}
	return ""
}

func getDurationInMillseconds(start time.Time) float64 {
	end := time.Now()
	duration := end.Sub(start)
	milliseconds := float64(duration) / float64(time.Millisecond)
	rounded := float64(int(milliseconds*100+.5)) / 100
	return rounded
}

func extractTokenFromHeader(r *http.Request) string {
	bearToken := r.Header.Get(JWTAuthHeader)
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}
