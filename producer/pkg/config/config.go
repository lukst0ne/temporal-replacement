package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisConfig *RedisConfig
	ASMConfig *ASMConfig
	DeviceManagementConfig *DeviceManagementConfig
}

type RedisConfig struct {
	Host string
	Port string
	DefaultQueueName string
	PriorityQueueName string
}

type ASMConfig struct {
	Host string
	PORT string
}

type DeviceManagementConfig struct {
	Host string
	Port string
}

func LoadFromEnv() *Config {
	err := godotenv.Load()
	if err != nil {
	  panic(err)
	}

    config := &Config{}
    
    if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
        config.RedisConfig.Host = redisHost
    }
    
	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		config.RedisConfig.Port = redisPort
	}
    
    return config
}