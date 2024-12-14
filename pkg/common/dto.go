package common

import (
	"errors"
)

var (
	ErrorInvalidParam = errors.New("invalid parameter")
	ErrorServer       = errors.New("server error")
	ErrorUnauthorized = errors.New("unauthorized")
)

// ErrorResponse is the error response type
type ErrorResponse struct {
	Message string `json:"message"`
}

// SuccessMessage is the success response type
type SuccessMessage struct {
	Message string `json:"message" example:"ok"`
}

// OkMsg is the default success response for 200 status code
var OkMessage SuccessMessage = SuccessMessage{
	Message: "ok",
}
