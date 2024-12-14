package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Cassandra *CassandraConfig `mapstructure:"cassandra"`
	Chat      *ChatConfig      `mapstructure:"chat"`
	Forwarder *ForwarderConfig `mapstructure:"forwarder"`
	Match     *MatchConfig     `mapstructure:"match"`
	Kafka     *KafkaConfig     `mapstructure:"kafka"`
	Redis     *RedisConfig     `mapstructure:"redis"`
	Uploader  *UploaderConfig  `mapstructure:"uploader"`
	Users     *UsersConfig     `mapstructure:"users"`
}

func NewConfig() (*Config, error) {
	setDefault()

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("failed to unmarshal configuration: %v", err)
		return nil, err
	}

	return &config, nil
}

func setDefault() {
	SetDefaultCassandraConfig()
	SetDefaultChatConfig()
	SetDefaultForwarderConfig()
	SetDefaultMatchConfig()
	SetDefaultKafkaConfig()
	SetDefaultRedisConfig()
	SetDefaultUploaderConfig()
	SetDefaultUserConfig()
}
