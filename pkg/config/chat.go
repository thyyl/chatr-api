package config

import (
	"os"

	"github.com/spf13/viper"
)

type ChatConfig struct {
	Http struct {
		Server struct {
			Port    string
			MaxConn int64
			Swag    bool
		}
	}
	Grpc struct {
		Server struct {
			Port string
		}
		Client struct {
			User struct {
				Endpoint string
			}
			Forwarder struct {
				Endpoint string
			}
		}
	}
	Subscriber struct {
		Id string
	}
	Message struct {
		MaxNum        int64
		PaginationNum int
		MaxSizeByte   int64
	}
	JWT struct {
		Secret           string
		ExpirationSecond int64
	}
}

func SetDefaultChatConfig() {
	viper.SetDefault("chat.http.server.port", "80")
	viper.SetDefault("chat.http.server.maxConn", 200)
	viper.SetDefault("chat.http.server.swag", false)
	viper.SetDefault("chat.grpc.server.port", "4000")
	viper.SetDefault("chat.grpc.client.user.endpoint", "reverse-proxy:80")
	viper.SetDefault("chat.grpc.client.forwarder.endpoint", "reverse-proxy:80")
	viper.SetDefault("chat.subscriber.id", "rc.msg."+os.Getenv("HOSTNAME"))
	viper.SetDefault("chat.message.maxNum", 5000)
	viper.SetDefault("chat.message.paginationNum", 5000)
	viper.SetDefault("chat.message.maxSizeByte", 4096)
	viper.SetDefault("chat.jwt.secret", "mysecret")
	viper.SetDefault("chat.jwt.expirationSecond", 86400)
}
