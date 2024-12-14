package uploader

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"github.com/google/uuid"
)

func newObjectKey(channelId uint64, extension string) string {
	return joinStrings(strconv.FormatUint(channelId, 10), "/", uuid.New().String(), extension)
}

func getChannelIdFromObjectKey(objectKey string) (uint64, error) {
	channelIdString := strings.Split(objectKey, "/")[0]
	channelId, err := strconv.ParseUint(channelIdString, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse channel ID from object key: %v, error: %v", objectKey, err)
	}
	return channelId, nil
}

func joinStrings(values ...string) string {
	var stringBuilder strings.Builder
	for _, value := range values {
		stringBuilder.WriteString(value)
	}

	return stringBuilder.String()
}

func byteSlice2String(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}
