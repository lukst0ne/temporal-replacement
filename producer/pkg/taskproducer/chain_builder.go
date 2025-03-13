package taskproducer

import (
	"fmt"

	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/google/uuid"
)

type ServiceRequest struct {
	ServiceName string
	DeviceId string
	Metadata map[string]interface{}
}

func (p *Producer) BuildChain(deviceId string, services []ServiceRequest, withPriority bool) *tasks.Chain {
	chainUUID := fmt.Sprintf("task:%s:%s", deviceId, uuid.New().String()[:8])
	routingKey := p.Config.RedisConfig.DefaultQueueName
	if withPriority {
		routingKey = p.Config.RedisConfig.PriorityQueueName
	}

	allTasks := make([]*tasks.Signature, len(services) + 4)

	allTasks = append(allTasks, generateAcquireLockTask(chainUUID, routingKey, deviceId))


	// getStatusTask := tasks.Signature{
	// 	UUID: fmt.Sprintf("%s:%s", chainUUID, "getStatus"),
	// 	RoutingKey: routingKey,
	// 	Name: "getStatus",
	// 	Args: []tasks.Arg{
	// 		{
	// 			Name: "deviceId",
	// 			Type:  "string",
	// 			Value: deviceId,
	// 		},
	// 	},
	// }

	// var enableServiceTasks []*tasks.Signature
	// for i, service := range(services) {
	// 	enableServiceTask := tasks.Signature{
	// 		UUID: fmt.Sprintf("%s:%s", chainUUID, fmt.Sprintf("enableService:%s", service.ServiceName)),
	// 		RoutingKey: routingKey,
	// 		Name: "enableService",
	// 	}
	// 	enableServiceTasks = append(enableServiceTasks, &enableServiceTask)

	// }
	
	
	return &tasks.Chain{Tasks: allTasks}
}

func generateAcquireLockTask(chainUUID string, routingKey string, deviceId string) *tasks.Signature {
	return &tasks.Signature{
		UUID: fmt.Sprintf("%s:%s", chainUUID, "acquireDeviceLock"),
		RoutingKey: routingKey,
		Name: "acquireDeviceLock",
		Args: []tasks.Arg{
			{
				Name: "deviceId",
				Type:  "string",
				Value: deviceId,
			},
		},
	}
}
