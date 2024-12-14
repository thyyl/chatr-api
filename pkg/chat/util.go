package chat

import (
	"encoding/json"
)

func DecodeToMessageDto(data []byte) (*MessageDto, error) {
	var msg MessageDto
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func DecodeToMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
