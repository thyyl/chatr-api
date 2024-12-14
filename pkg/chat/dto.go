package chat

import (
	"encoding/json"
	"strconv"

	"github.com/thyyl/chatr/pkg/common"
)

type MessageDto struct {
	MessageId string `json:"messageId"`
	Event     int    `json:"event"`
	UserId    string `json:"userId"`
	Payload   string `json:"payload"`
	Seen      bool   `json:"seen"`
	Time      int64  `json:"time"`
}

type MessagesDto struct {
	NextPageState string       `json:"nextPageState"`
	Messages      []MessageDto `json:"messages"`
}

type UserDto struct {
	Id   string `json:"id"`
	Name string `json:"name" binding:"required"`
}

type UserIdsDto struct {
	UserIds []string `json:"userIds"`
}

func (m *MessageDto) Encode() []byte {
	result, _ := json.Marshal(m)
	return result
}

func (m *MessageDto) ToMessage(accessToken string) (*Message, error) {
	authResult, err := common.Auth(&common.AuthPayload{
		AccessToken: accessToken,
	})
	if err != nil {
		return nil, err
	}
	if authResult.Expired {
		return nil, common.ErrorTokenExpired
	}
	channelID := authResult.ChannelId
	userID, err := strconv.ParseUint(m.UserId, 10, 64)
	if err != nil {
		return nil, err
	}
	return &Message{
		Event:     m.Event,
		ChannelId: channelID,
		UserId:    userID,
		Payload:   m.Payload,
		Time:      m.Time,
	}, nil
}
