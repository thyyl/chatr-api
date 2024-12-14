package common

import "errors"

var (
	ErrorUserNotFound           = errors.New("error user not found")
	ErrorSessionNotFound        = errors.New("error session not found")
	ErrorOpenFile               = errors.New("fail to open file")
	ErrorReceiveFile            = errors.New("no file is received")
	ErrorUploadFile             = errors.New("fail to upload file")
	ErrorTooManyUploads         = errors.New("too many uploads")
	ErrorChannelOrUserNotFound  = errors.New("error channel or user not found")
	ErrorExceedMessageNumLimits = errors.New("error exceed max number of messages")
)
