package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/RichardKnop/machinery/v2"
	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/google/uuid"

	"github.com/lukst0ne/temporal-replacement/consumer/pkg/workflows"

	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	redislock "github.com/RichardKnop/machinery/v2/locks/redis"
)

var REDIS_HOST = os.Getenv("REDIS_HOST")

func main() {
	server, _ := setupProducerServer()

	// var count int
	for i := 0; i < 20; i++ {
		uuid := uuid.New().String()
		deviceId := fmt.Sprintf("DEVICE_%d", i%10)
		workflows := []*tasks.Signature{
			{
				UUID: fmt.Sprintf("task:%s:%s:lock", deviceId, uuid),
				Name: "acquireDeviceLock",
				Args: []tasks.Arg{
					{
						Name:  "deviceId",
						Type:  "string",
						Value: deviceId,
					},
				},
				RetryCount: 5,
			},
			{
				UUID: fmt.Sprintf("task:%s:%s:getStatuses", deviceId, uuid),
				Name: "getAllServiceStatus",
				Args: []tasks.Arg{
					{
						Name:  "deviceId",
						Type:  "string",
						Value: deviceId,
					},
				},
			},
			{
				UUID: fmt.Sprintf("task:%s:%s:startup", deviceId, uuid),
				Name: "enableService",
				Args: []tasks.Arg{
					{
						Name:  "deviceId",
						Type:  "string",
						Value: deviceId,
					},
					{
						Name:  "service",
						Type:  "string",
						Value: "startup",
					},
				},
				RetryCount: 2,
			},
			{
				UUID: fmt.Sprintf("task:%s:%s:releaseLock", deviceId, uuid),
				Name: "releaseDeviceLock",
				Args: []tasks.Arg{
					{
						Name:  "deviceId",
						Type:  "string",
						Value: deviceId,
					},
				},
			},
		}

		fmt.Printf("pushing task:%s:%s:\n", deviceId, uuid)
		chain, _ := tasks.NewChain(workflows...)

		server.SendChainWithContext(context.TODO(), chain)

		time.Sleep(10 * time.Millisecond)
		// count++
	}
}

// Define a simple task function
func setupProducerServer() (*machinery.Server, error) {
	cnf := &config.Config{
		DefaultQueue:    "machinery_tasks",
		ResultsExpireIn: 3600,
		Redis: &config.RedisConfig{
			MaxIdle:                3,
			IdleTimeout:            240,
			ReadTimeout:            15,
			WriteTimeout:           15,
			ConnectTimeout:         15,
			NormalTasksPollPeriod:  1000,
			DelayedTasksPollPeriod: 500,
		},
	}

	broker := redisbroker.NewGR(cnf, []string{fmt.Sprintf("%s:6379", REDIS_HOST)}, 0)
	backend := redisbackend.NewGR(cnf, []string{fmt.Sprintf("%s:6379", REDIS_HOST)}, 0)
	lock := redislock.New(cnf, []string{fmt.Sprintf("%s:6379", REDIS_HOST)}, 0, 3)
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
