package config

import "github.com/spf13/viper"

type KafkaConfig struct {
	Address string
	Version string
}

func SetDefaultKafkaConfig() {
	viper.SetDefault("kafka.address", "kafka:9092")
	viper.SetDefault("kafka.version", "3.6.0")
}
