package chat

import (
	"encoding/json"
	"strconv"
)

const (
	EventText = iota
	EventAction
	EventSeen
	EventFile
)

type Action string

var (
	WaitingMessage   Action = "waiting"
	JoinedMessage    Action = "joined"
	IsTypingMessage  Action = "istyping"
	EndTypingMessage Action = "endtyping"
	OfflineMessage   Action = "offline"
	LeavedMessage    Action = "leaved"
)

type Message struct {
	MessageId uint64 `json:"messageId"`
	Event     int    `json:"event"`
	ChannelId uint64 `json:"channelId"`
	UserId    uint64 `json:"userId"`
	Payload   string `json:"payload"`
	Seen      bool   `json:"seen"`
	Time      int64  `json:"time"`
}

type Channel struct {
	Id          uint64 `json:"id"`
	AccessToken string `json:"accessToken"`
}

type User struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

func (m *Message) Encode() []byte {
	result, _ := json.Marshal(m)
	return result
}

func (m *Message) ToPresenter() *MessageDto {
	return &MessageDto{
		MessageId: strconv.FormatUint(m.MessageId, 10),
		Event:     m.Event,
		UserId:    strconv.FormatUint(m.UserId, 10),
		Payload:   m.Payload,
		Seen:      m.Seen,
		Time:      m.Time,
	}
}
