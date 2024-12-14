package chat

import "github.com/thyyl/chatr/pkg/infra"

type InfraCloser struct{}

func NewInfraCloser() *InfraCloser {
	return &InfraCloser{}
}

func (closer *InfraCloser) Close() error {
	if err := ForwarderConn.Conn.Close(); err != nil {
		return err
	}

	if err := UserConn.Conn.Close(); err != nil {
		return err
	}

	infra.CassandraSession.Close()
	return infra.RedisClient.Close()
}
