package config

import "github.com/spf13/viper"

type MatchConfig struct {
	Http struct {
		Server struct {
			Port    string
			MaxConn int64
			Swag    bool
		}
	}
	Grpc struct {
		Client struct {
			Chat struct {
				Endpoint string
			}
			User struct {
				Endpoint string
			}
		}
	}
}

func SetDefaultMatchConfig() {
	viper.SetDefault("match.http.server.port", "5002")
	viper.SetDefault("match.http.server.maxConn", 200)
	viper.SetDefault("match.http.server.swag", false)
	viper.SetDefault("match.grpc.client.chat.endpoint", "reverse-proxy:80")
	viper.SetDefault("match.grpc.client.user.endpoint", "reverse-proxy:80")
}
