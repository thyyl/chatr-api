package forwarder

import "github.com/thyyl/chatr/pkg/infra"

type InfraCloser struct{}

func NewInfraCloser() *InfraCloser {
	return &InfraCloser{}
}

func (c *InfraCloser) Close() error {
	return infra.RedisClient.Close()
}
