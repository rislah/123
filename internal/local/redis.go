package local

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/cep21/circuit/v3"
	"github.com/rislah/fakes/internal/redis"
)

func NewRedis() (redis.Client, error) {
	srv, err := miniredis.Run()
	if err != nil {
		return nil, err
	}

	rc, err := redis.NewClient(srv.Addr(), &circuit.Circuit{}, nil)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func MakeRedis() (redis.Client, func() error, error) {
	redis, err := NewRedis()
	return redis, redis.FlushAll, err
}
