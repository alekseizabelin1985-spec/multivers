// Package main is the entry point for BanOfWorld.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"multiverse-core/internal/eventbus"
	"multiverse-core/services/banofworld"
)

func main() {
	// Initialize event bus
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"redpanda:9092"}
	}
	bus := eventbus.NewEventBus(brokers)

	// Create and run service
	service := banofworld.NewService(bus)

	// Handle shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down BanOfWorld...")
		cancel()
	}()

	log.Println("BanOfWorld starting...")
	if err := service.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal("Service failed:", err)
	}
	log.Println("BanOfWorld stopped.")
}
