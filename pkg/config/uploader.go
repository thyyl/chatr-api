package config

import "github.com/spf13/viper"

type RateLimitConfig struct {
	RatePerSecond int
	Burst         int
}

type UploaderConfig struct {
	Http struct {
		Server struct {
			Port          string
			Swag          bool
			MaxMemoryByte int64
			MaxBodyByte   int64
		}
	}
	S3 struct {
		Endpoint              string
		Region                string
		Bucket                string
		AccessKey             string
		SecretKey             string
		PresignLifetimeSecond int64
	}
	RateLimit struct {
		ChannelUpload RateLimitConfig
	}
}

func SetDefaultUploaderConfig() {
	viper.SetDefault("uploader.http.server.port", "5003")
	viper.SetDefault("uploader.http.server.swag", false)
	viper.SetDefault("uploader.http.server.maxBodyByte", "67108864")   // 64MB
	viper.SetDefault("uploader.http.server.maxMemoryByte", "16777216") // 16MB
	viper.SetDefault("uploader.s3.endpoint", "http://localhost:9000")
	viper.SetDefault("uploader.s3.region", "us-east-1")
	viper.SetDefault("uploader.s3.bucket", "myfilebucket")
	viper.SetDefault("uploader.s3.accessKey", "")
	viper.SetDefault("uploader.s3.secretKey", "")
	viper.SetDefault("uploader.s3.presignLifetimeSecond", 86400)
	viper.SetDefault("uploader.rateLimit.channelUpload.rps", 200)
	viper.SetDefault("uploader.rateLimit.channelUpload.burst", 50)
}
