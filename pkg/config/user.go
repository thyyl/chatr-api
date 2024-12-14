package config

import (
	"github.com/spf13/viper"
)

type CookieConfig struct {
	MaxAge int
	Path   string
	Domain string
}

type UsersConfig struct {
	Http struct {
		Server struct {
			Port string
			Swag bool
		}
	}
	Grpc struct {
		Server struct {
			Port string
		}
	}
	OAuth struct {
		Cookie CookieConfig
		Google struct {
			RedirectUrl  string
			ClientId     string
			ClientSecret string
			Scopes       string
		}
	}
	Auth struct {
		Cookie CookieConfig
	}
}

func SetDefaultUserConfig() {
	viper.SetDefault("users.http.server.port", "80")
	viper.SetDefault("users.http.server.swag", false)
	viper.SetDefault("users.grpc.server.port", "4000")
	viper.SetDefault("users.oauth.cookie.maxAge", 3600)
	viper.SetDefault("users.oauth.cookie.path", "/")
	viper.SetDefault("users.oauth.cookie.domain", "localhost")
	viper.SetDefault("users.oauth.google.redirectUrl", "http://localhost/api/user/oauth2/google/callback")
	viper.SetDefault("users.oauth.google.clientId", "")
	viper.SetDefault("users.oauth.google.clientSecret", "")
	viper.SetDefault("users.oauth.google.scopes", "https://www.googleapis.com/auth/userinfo.email,https://www.googleapis.com/auth/userinfo.profile")
	viper.SetDefault("users.auth.cookie.maxAge", 86400)
	viper.SetDefault("users.auth.cookie.path", "/")
	viper.SetDefault("users.auth.cookie.domain", "localhost")
}
