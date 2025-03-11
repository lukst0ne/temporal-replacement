package workflows

import (
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type WorkerEnv struct {
	Redis *redis.Client
	Logger *logrus.Logger
}

func NewWorker(addr string, logger *logrus.Logger, db int) (*WorkerEnv, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})
	workerEnv := &WorkerEnv{
		Redis: rdb,
		Logger: logger,
	}

	return workerEnv, nil
}
