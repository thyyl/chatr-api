package config

import "github.com/spf13/viper"

type RedisConfig struct {
	Password                 string
	Address                  string
	ExpirationHours          int64
	MinIdleConnections       int
	PoolSize                 int
	ReadTimeoutMilliSeconds  int
	WriteTimeoutMilliSeconds int
}

func SetDefaultRedisConfig() {
	viper.SetDefault("redis.password", "pass.123")
	viper.SetDefault("redis.address", "localhost:6379")
	viper.SetDefault("redis.expirationHours", 24)
	viper.SetDefault("redis.minIdleConnection", 16)
	viper.SetDefault("redis.poolSize", 64)
	viper.SetDefault("redis.readTimeoutMilliSecond", 3000)
	viper.SetDefault("redis.writeTimeoutMilliSecond", 3000)
}
