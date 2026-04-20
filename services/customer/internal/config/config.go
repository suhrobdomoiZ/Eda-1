package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPC     GRPCConfig
	Postgres PostgresConfig
}

type GRPCConfig struct {
	Port string
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.DBName,
	)
}

func Load() *Config {
	cfg := &Config{
		GRPC: GRPCConfig{
			Port: getEnv("CUSTOMER_GRPC_PORT", "50051"),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("DATABASE_HOST", "localhost"),
			Port:     getEnv("DATABASE_PORT", "5432"),
			User:     getEnv("DATABASE_USER", ""),
			Password: getEnv("DATABASE_PASSWORD", ""),
			DBName:   getEnv("DATABASE_NAME", ""),
		},
	}
	if cfg.Postgres.User == "" || cfg.Postgres.DBName == "" {
		panic("DATABASE_USER and DATABASE_NAME are required")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
