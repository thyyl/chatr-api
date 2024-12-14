package config

import "github.com/spf13/viper"

type CassandraConfig struct {
	Hosts    string
	Port     int
	User     string
	Password string
	Keyspace string
}

func SetDefaultCassandraConfig() {
	viper.SetDefault("cassandra.hosts", "localhost")
	viper.SetDefault("cassandra.port", 9042)
	viper.SetDefault("cassandra.user", "")
	viper.SetDefault("cassandra.password", "")
	viper.SetDefault("cassandra.keyspace", "chatr")
}
