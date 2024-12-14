package uploader

import (
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/config"
)

type ChannelUploadRateLimiter struct {
	*common.RateLimiter
}

func NewChannelUploadRateLimiter(rc redis.UniversalClient, config *config.Config) ChannelUploadRateLimiter {
	return ChannelUploadRateLimiter{
		common.NewRateLimiter(
			rc,
			config.Uploader.RateLimit.ChannelUpload.RatePerSecond,
			config.Uploader.RateLimit.ChannelUpload.Burst,
			time.Duration(config.Redis.ExpirationHours)*time.Hour,
		),
	}
}
