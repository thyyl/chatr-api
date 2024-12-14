package infra

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/config"
)

var (
	logger = watermill.NewStdLogger(false, false)
)

func NewKafkaPublisher(config *config.Config) (message.Publisher, error) {
	kafkaPublisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   common.GetServerAddress(config.Kafka.Address),
			Marshaler: kafka.DefaultMarshaler{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return kafkaPublisher, nil
}

func NewKafkaSubscriber(config *config.Config) (message.Subscriber, error) {
	saramaConfig := sarama.NewConfig()
	saramaVersion, err := sarama.ParseKafkaVersion(config.Kafka.Version)
	if err != nil {
		return nil, err
	}

	saramaConfig.Version = saramaVersion
	saramaConfig.Consumer.Fetch.Default = 1024 * 1024
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = true
	saramaConfig.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	kafkaSubscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:                common.GetServerAddress(config.Kafka.Address),
			Unmarshaler:            kafka.DefaultMarshaler{},
			ConsumerGroup:          watermill.NewUUID(),
			InitializeTopicDetails: &sarama.TopicDetail{NumPartitions: 1, ReplicationFactor: 2},
			OverwriteSaramaConfig:  saramaConfig,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return kafkaSubscriber, nil
}

func NewBrokerRouter(name string) (*message.Router, error) {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	router.AddMiddleware(
		middleware.CorrelationID,
		middleware.Timeout(time.Second*15),
		middleware.Recoverer,
	)
	return router, nil
}
