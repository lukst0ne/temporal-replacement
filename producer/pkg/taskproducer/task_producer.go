package taskproducer

import (
	"fmt"
	"temporal-replacement/producer/pkg/config"

	"github.com/RichardKnop/machinery/v2"
	mc "github.com/RichardKnop/machinery/v2/config"

	"github.com/lukst0ne/temporal-replacement/consumer/pkg/workflows"

	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	redislock "github.com/RichardKnop/machinery/v2/locks/redis"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type Producer struct {
	Config          *config.Config
	Logger          Logger
	MachineryServer *machinery.Server
}

func NewProducer(config *config.Config, logger Logger) (*Producer, error) {
	server, err := initMachineryServer(config)
	if err != nil {
		logger.Errorf("Failed to set up server: %s", err)
		return nil, err
	}
	return &Producer{
		Config:          config,
		Logger:          logger,
		MachineryServer: server,
	}, nil
}

func initMachineryServer(config *config.Config) (*machinery.Server, error) {
	cnf := &mc.Config{
		DefaultQueue:    config.RedisConfig.DefaultQueueName,
		ResultsExpireIn: 3600,
		Redis: &mc.RedisConfig{
			MaxIdle:                3,
			IdleTimeout:            240,
			ReadTimeout:            15,
			WriteTimeout:           15,
			ConnectTimeout:         15,
			NormalTasksPollPeriod:  1000,
			DelayedTasksPollPeriod: 500,
		},
	}

	addrs := []string{fmt.Sprintf("%s:%s", config.RedisConfig.Host, config.RedisConfig.Port)}

	broker := redisbroker.NewGR(cnf, addrs, 0)
	backend := redisbackend.NewGR(cnf, addrs, 0)
	lock := redislock.New(cnf, addrs, 0, 3)
	server := machinery.NewServer(cnf, broker, backend, lock)

	workerEnv := &workflows.WorkerEnv{}
	workflowsMap := map[string]interface{}{
		"acquireDeviceLock":   workerEnv.AcquireDeviceLock,
		"getAllServiceStatus": workerEnv.GetAllServiceStatus,
		"enableService":       workerEnv.EnableService,
		"releaseDeviceLock":   workerEnv.ReleaseDeviceLock,
	}

	err := server.RegisterTasks(workflowsMap)
	if err != nil {
		fmt.Printf("error registering tasks: %s", err.Error())
		return nil, err
	}
	return server, nil
}
