package main

import (
	"temporal-replacement/producer/pkg/config"
	"temporal-replacement/producer/pkg/taskproducer"

	"go.uber.org/zap"
	// "github.com/lukst0ne/temporal-replacement/consumer/pkg/workflows"
)

func main() {
	config := config.LoadFromEnv()

	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	
	producer, err := taskproducer.NewProducer(config, sugar)
	if err != nil {
		panic(err)
	}

    services := []taskproducer.ServiceRequest{
        {
            ServiceName: "startup",
            DeviceId:    "device1",
            Metadata:    nil,
        },
        {
            ServiceName: "greTunnel",
            DeviceId:    "device1",
            Metadata: map[string]interface{}{
				"tunnels": []map[string]interface{}{
					{
						"tunnelId": 1,
						"remoteIp": "",
					},
					{
						"tunnelId": 2,
						"remoteIp": "",
					},
				},
            },
        },
    }


	chain := producer.BuildChain("device1", services, false)

	producer.MachineryServer.SendChain(chain)
	

}
