package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTP        HTTP
	Postgres    Postgres
	Redis       Redis
	JWT         JWT
	Environment string
}

type HTTP struct {
	Port string
}

type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type Redis struct {
	URL string
	TTL time.Duration
}

type JWT struct {
	Secret string
}

var cfg *Config

func Load() *Config {
	if cfg != nil {
		return cfg
	}

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	cfg = &Config{
		HTTP: HTTP{
			Port: getEnv("HTTP_PORT", "8080"),
		},
		Postgres: Postgres{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "task_manager"),
		},
		Redis: Redis{
			URL: getEnv("REDIS_URL", "redis://localhost:6379"),
			TTL: parseDuration(getEnv("REDIS_TTL", "1h")),
		},
		JWT: JWT{
			Secret: getEnv("JWT_SECRET", "super-secret-key"),
		},
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	return cfg
}

// Вспомогательные функции
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func parseDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalf("Invalid duration format for %s: %v", value, err)
	}
	return duration
}
