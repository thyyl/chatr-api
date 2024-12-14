package common

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sony/sonyflake"
)

type IDGenerator interface {
	NextID() (uint64, error)
}

func NewSonyFlake() (IDGenerator, error) {
	var settings sonyflake.Settings
	sf := sonyflake.NewSonyflake(settings)
	if sf == nil {
		return nil, errors.New("sonyflake not created")
	}
	return sf, nil
}

func GetServerAddress(addrs string) []string {
	return strings.Split(addrs, ",")
}

func Join(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}

func Response(c *gin.Context, httpCode int, err error) {
	message := err.Error()

	c.JSON(httpCode, ErrorResponse{
		Message: message,
	})
}

func Encode[T any](v T) []byte {
	result, err := json.Marshal(v)
	if err != nil {
		log.Printf("Failed to marshal: %v", err)
		return nil
	}
	return result
}
