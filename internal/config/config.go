package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost          string
	DBPort          int
	DBUser          string
	DBPassword      string
	DBName          string
	NATSClusterID   string
	NATSClientID    string
	NATSChannel     string
	NATSDurableID   string
	HTTPPort        string
}

func Load() *Config {
	return &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnvAsInt("DB_PORT", 5433),
		DBUser:        getEnv("DB_USER", "myuser"),
		DBPassword:    getEnv("DB_PASSWORD", "mypassword"),
		DBName:        getEnv("DB_NAME", "myapp"),
		NATSClusterID: getEnv("NATS_CLUSTER_ID", "my-cluster"),
		NATSClientID:  getEnv("NATS_CLIENT_ID", "order-service"),
		NATSChannel:   getEnv("NATS_CHANNEL", "orders"),
		NATSDurableID: getEnv("NATS_DURABLE_ID", "order-service-durable"),
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}