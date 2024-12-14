package chat

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/thyyl/chatr/pkg/common"
	"github.com/thyyl/chatr/pkg/config"
	"gopkg.in/olahol/melody.v1"
)

type MessageSubscriber struct {
	subscriberId   string
	router         *message.Router
	subscriber     message.Subscriber
	melodyChatConn MelodyChatConn
}

func NewMessageSubscriber(name string, router *message.Router, config *config.Config, subscriber message.Subscriber, melodyChatConn MelodyChatConn) (*MessageSubscriber, error) {
	subscriberId := config.Chat.Subscriber.Id

	return &MessageSubscriber{
		subscriberId:   subscriberId,
		router:         router,
		subscriber:     subscriber,
		melodyChatConn: melodyChatConn,
	}, nil
}

func (s *MessageSubscriber) HandleMessage(chatMessage *message.Message) error {
	message, err := DecodeToMessage([]byte(chatMessage.Payload))
	if err != nil {
		return err
	}

	return s.sendMessage(context.Background(), message)
}

func (s *MessageSubscriber) RegisterHandler() {
	s.router.AddNoPublisherHandler(
		"chatr_message_handler",
		s.subscriberId,
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

func (s *MessageSubscriber) sendMessage(ctx context.Context, message *Message) error {
	return s.melodyChatConn.BroadcastFilter(message.ToPresenter().Encode(), func(session *melody.Session) bool {
		channelId, exist := session.Get(common.SessionCidKey)
		if !exist {
			return false
		}

		return message.ChannelId == (channelId.(uint64))
	})
}
