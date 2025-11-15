// cmd/narrative-orchestrator/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"multiverse-core/services/narrativeorchestrator"
)

func main() {
	cfg := narrativeorchestrator.Config{
		KafkaBrokers: getEnvBrokers("KAFKA_BROKERS", []string{"redpanda:9092"}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down NarrativeOrchestrator...")
		cancel()
	}()

	service, err := narrativeorchestrator.NewService(cfg)
	if err != nil {
		log.Fatal("Failed to initialize NarrativeOrchestrator:", err)
	}

	service.Start(ctx)
	<-ctx.Done()
	service.Stop()
	log.Println("NarrativeOrchestrator stopped.")
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
