package match

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/thyyl/chatr/pkg/common"
	"gopkg.in/olahol/melody.v1"
)

type MatchSubscriber struct {
	melodyMatch MelodyMatchConn
	router      *message.Router
	userService UserService
	subscriber  message.Subscriber
}

func NewMatchSubscriber(name string, melodyMatch MelodyMatchConn, router *message.Router, userService UserService, subscriber message.Subscriber) *MatchSubscriber {
	return &MatchSubscriber{
		melodyMatch: melodyMatch,
		router:      router,
		userService: userService,
		subscriber:  subscriber,
	}
}

func (s *MatchSubscriber) Run() error {
	return s.router.Run(context.Background())
}

func (s *MatchSubscriber) GracefulStop() error {
	return s.router.Close()
}

func (s *MatchSubscriber) RegisterHandler() {
	s.router.AddNoPublisherHandler(
		"chatr_match_result_handler",
		common.MatchPubSubTopicRcKey,
		s.subscriber,
		s.HandleMatchResult,
	)
}

func (s *MatchSubscriber) HandleMatchResult(message *message.Message) error {
	result, err := DecodeToMatchResult([]byte(message.Payload))
	if err != nil {
		return err
	}

	ctx := context.Background()
	return s.sendMatchResult(ctx, result)
}
func (s *MatchSubscriber) sendMatchResult(ctx context.Context, result *MatchResult) error {
	return s.melodyMatch.BroadcastFilter(result.ToDto().Encode(), func(session *melody.Session) bool {
		sessionUserId, exist := session.Get(common.SessionUidKey)
		if !exist {
			return false
		}

		userID := sessionUserId.(uint64)
		if (userID == result.PeerId) || (userID == result.UserId) {
			if err := s.userService.AddUserToChannel(ctx, result.ChannelId, userID); err != nil {
				slog.Error(err.Error())
				return false
			}
			return true
		}
		return false
	})
}
