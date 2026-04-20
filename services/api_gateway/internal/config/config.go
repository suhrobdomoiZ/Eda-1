package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTP     HTTPConfig
	Services ServicesConfig
}

type HTTPConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type ServicesConfig struct {
	AuthAddr       string
	RestaurantAddr string
	CustomerAddr   string
	CourierAddr    string
}

func Load() *Config {
	cfg := &Config{
		HTTP: HTTPConfig{
			Port:         getEnv("GATEWAY_HTTP_PORT", "8080"),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Services: ServicesConfig{
			AuthAddr:       getEnv("AUTH_GRPC_ADDR", ""),
			RestaurantAddr: getEnv("RESTAURANT_GRPC_ADDR", ""),
			CustomerAddr:   getEnv("CUSTOMER_GRPC_ADDR", ""),
			CourierAddr:    getEnv("COURIER_GRPC_ADDR", ""),
		},
	}

	if cfg.Services.AuthAddr == "" {
		panic("AUTH_GRPC_ADDR is required")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (h HTTPConfig) Addr() string {
	return fmt.Sprintf(":%s", h.Port)
}
