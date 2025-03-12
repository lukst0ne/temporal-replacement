package main

import (
	"fmt"
	"os"
	"temporal-replacement/consumer/pkg/workflows"

	"github.com/RichardKnop/machinery/v2"
	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/sirupsen/logrus"

	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	redislock "github.com/RichardKnop/machinery/v2/locks/redis"
)

var REDIS_HOST = os.Getenv("REDIS_HOST")

func main() {
	server, err := setupServer()
	if err != nil {
		log.INFO.Fatal("Failed to set up server: %s", err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	log.Set(logger)

	workerEnv, _ := workflows.NewWorker(fmt.Sprintf("%s:6379", REDIS_HOST), logger, 0)

	workflowsMap := map[string]interface{}{
		"acquireDeviceLock":   workerEnv.AcquireDeviceLock,
		"getAllServiceStatus": workerEnv.GetAllServiceStatus,
		"enableService":       workerEnv.EnableService,
		"releaseDeviceLock":   workerEnv.ReleaseDeviceLock,
	}

	err = server.RegisterTasks(workflowsMap)
	if err != nil {
		fmt.Printf("error registering tasks: %s", err.Error())
	}

	machineryWorker := server.NewWorker("worker_name", 1)
	machineryPriorityWorker := server.NewCustomQueueWorker("priority_worker_name", 10, "asm_priority_queue")

	machineryWorker.SetPostTaskHandler(workerEnv.PrintServiceStatus)
	machineryPriorityWorker.SetPostTaskHandler(workerEnv.PrintServiceStatus)

	log.WARNING.Println("Starting worker")

	errorsChan := make(chan error)
	machineryWorker.LaunchAsync(errorsChan)
	machineryPriorityWorker.LaunchAsync(errorsChan)

	log.INFO.Println("Workers started successfully")
	select {}
}

func setupServer() (*machinery.Server, error) {
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

	return server, nil
}
