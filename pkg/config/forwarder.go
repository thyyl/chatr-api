package config

import (
	"github.com/spf13/viper"
)

type ForwarderConfig struct {
	Grpc struct {
		Server struct {
			Port string
		}
	}
}

func SetDefaultForwarderConfig() {
	viper.SetDefault("forwarder.grpc.server.port", "4000")
}
