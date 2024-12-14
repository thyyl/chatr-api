package forwarder

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/thyyl/chatr/pkg/chat"
	"github.com/thyyl/chatr/pkg/common"
)

type MessageSubscriber struct {
	router           *message.Router
	subscriber       message.Subscriber
	forwarderService ForwarderService
}

func NewMessageSubscriber(router *message.Router, subscriber message.Subscriber, forwarderService ForwarderService) *MessageSubscriber {
	return &MessageSubscriber{
		router:           router,
		subscriber:       subscriber,
		forwarderService: forwarderService,
	}
}

func (s *MessageSubscriber) HandleMessage(message *message.Message) error {
	payload, err := chat.DecodeToMessage([]byte(message.Payload))
	if err != nil {
		return err
	}

	return s.forwarderService.ForwardMessage(message.Context(), payload)
}

func (s *MessageSubscriber) RegisterHandler() {
	s.router.AddNoPublisherHandler(
		"chatr_message_forwarder",
		common.MessagePubTopic,
		s.subscriber,
		s.HandleMessage,
	)
}

func (s *MessageSubscriber) Run() error {
	return s.router.Run(context.Background())
}

func (s *MessageSubscriber) GracefulStop() error {
	return s.router.Close()
}
