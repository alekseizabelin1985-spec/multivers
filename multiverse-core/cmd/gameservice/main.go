package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"multiverse-core/services/gameservice"
)

func main() {
	// Конфигурация из окружения
	cfg := gameservice.Config{
		KafkaBrokers: getEnvBrokers("KAFKA_BROKERS", []string{"redpanda:9092"}),
		HTTPAddr:     getEnv("HTTP_ADDR", ":8080"),
		CacheTTL:     getEnvDuration("CACHE_TTL", time.Minute*5),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down GameService...")
		cancel()
	}()

	// Запуск сервиса
	service := gameservice.NewService(cfg)
	service.Start(ctx)
	<-ctx.Done()
	service.Stop()
	log.Println("GameService stopped.")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvBrokers(key string, fallback []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		brokers := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				brokers = append(brokers, trimmed)
			}
		}
		if len(brokers) > 0 {
			return brokers
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}