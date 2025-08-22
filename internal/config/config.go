package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBConfig     DBConfig
	GRPCPort     string
	RESTPort     string
	WorkerCount  int
	DataInterval int // in milliseconds
	LogLevel     string
}

type DBConfig struct {
	DBDriver         string
	DBSource         string
	MaxDBConnections int
	MinDBConnections int
	MaxConnLifetime  time.Duration
	MaxConnIdleTime  time.Duration
}

func LoadConfig() *Config {
	return &Config{
		DBConfig: DBConfig{
			DBDriver: getEnv("DB_DRIVER", "postgres"),
			DBSource: getEnv("DB_SOURCE", "postgres://elena:testy@postgres:5432/data_aggregator?sslmode=disable"),

			MaxDBConnections: getEnvAsInt("MAX_DB_CONNECTIONS", 10),
			MinDBConnections: getEnvAsInt("MIN_DB_CONNECTIONS", 2),
			MaxConnLifetime:  time.Duration(getEnvAsInt("MAX_CONN_LIFETIME", 3600)) * time.Second,
			MaxConnIdleTime:  time.Duration(getEnvAsInt("MAX_CONN_IDLE_TIME", 1800)) * time.Second,
		},
		GRPCPort:     getEnv("GRPC_PORT", ":9090"),
		RESTPort:     getEnv("REST_PORT", ":8080"),
		WorkerCount:  getEnvAsInt("WORKER_COUNT", 5),
		DataInterval: getEnvAsInt("DATA_INTERVAL", 100),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}
