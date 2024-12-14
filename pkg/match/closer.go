package match

import "github.com/thyyl/chatr/pkg/infra"

type InfraCloser struct{}

func NewInfraCloser() *InfraCloser {
	return &InfraCloser{}
}

func (closer *InfraCloser) Close() error {
	if err := ChatConn.Conn.Close(); err != nil {
		return err
	}

	if err := infra.RedisClient.Close(); err != nil {
		return err
	}

	return infra.RedisClient.Close()
}
