package workflows

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/RichardKnop/machinery/v2/tasks"
)

type ServiceStatus struct {
	DeviceId              string    `json:"deviceId"`
	ServiceName           string    `json:"serviceName"`
	Status                string    `json:"status"`
	LastWorkflowStartTime time.Time `json:"lastWorkflowStartTime"`
	LastWorkflowEndTime   time.Time `json:"lastWorkflowEndTime"`
}

func (w *WorkerEnv) GetAllServiceStatus(deviceId string) error {	
	serviceStatus := []ServiceStatus{
		{
			DeviceId:              deviceId,
			ServiceName:           "startup",
			Status:                "Completed",
			LastWorkflowStartTime: time.Now().Add(-2 * time.Hour),
			LastWorkflowEndTime:   time.Now().Add(-1 * time.Hour),
		},
	}

	for _, service := range serviceStatus {
		// Create key prefixes
		keyPrefix := deviceId + ":services:" + service.ServiceName

		// Set status
		statusKey := keyPrefix + ":status"
		if err := w.Redis.Set(statusKey, service.Status, 0).Err(); err != nil {
			return err
		}

		// Set lastWorkflowStartTime
		startTimeKey := keyPrefix + ":lastWorkflowStartTime"
		if err := w.Redis.Set(startTimeKey, service.LastWorkflowStartTime.Format(time.RFC3339), 0).Err(); err != nil {
			return err
		}

		// Set lastWorkflowEndTime
		endTimeKey := keyPrefix + ":lastWorkflowEndTime"
		if err := w.Redis.Set(endTimeKey, service.LastWorkflowEndTime.Format(time.RFC3339), 0).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (w *WorkerEnv) PrintServiceStatus(task *tasks.Signature) {
	switch task.Name {
	case "enableService":
		parts := strings.Split(task.UUID, ":")
		if len(parts) < 3 {
			return
		}
		deviceId := parts[1]
		service := parts[3]

		keyPrefix := deviceId + ":services:" + service
		statusKey := keyPrefix + ":status"
		status, err := w.Redis.Get(statusKey).Result()
		if err != nil {
			return
		}
		if status == "Completed" {
			w.Logger.Warnf("%s status: %s\n", task.UUID, status)
		} else if status == "Failed" {
			w.Logger.Errorf("%s status: %s\n", task.UUID, status)
		}
		return
	}
}

func (w *WorkerEnv) EnableService(deviceId, service string) error {
	//get service status from redis
	keyPrefix := deviceId + ":services:" + service
	statusKey := keyPrefix + ":status"
	if err := w.Redis.Set(statusKey, "Running", 0).Err(); err != nil {
		return err
	}
	startTimeKey := keyPrefix + ":lastWorkflowStartTime"
	if err := w.Redis.Set(startTimeKey, time.Now().Format(time.RFC3339), 0).Err(); err != nil {
		return err
	}

	//TODO pull script from dynamodb

	//runDSLScript
	err := w.runDSLScript(deviceId, service)
	if err != nil {
		if err := w.Redis.Set(statusKey, "Failed", 0).Err(); err != nil {
			return err
		}
		//dont set error on a failed runDSLScript so the next task can run
		return nil
	}

	if err := w.Redis.Set(statusKey, "Completed", 0).Err(); err != nil {
		return err
	}
	endTimeKey := keyPrefix + ":lastWorkflowEndTime"
	if err := w.Redis.Set(endTimeKey, time.Now().Format(time.RFC3339), 0).Err(); err != nil {
		return err
	}
	return nil
}

func (w *WorkerEnv) runDSLScript(deviceId, service string) error {
	time.Sleep(time.Second * time.Duration(rand.Intn(5)+1))
	if deviceId == "DEVICE_5" && service == "startup" {
		return errors.New("failed to run DSL script")
	}
	return nil
}
