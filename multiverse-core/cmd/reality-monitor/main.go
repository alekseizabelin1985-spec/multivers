package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"multiverse-core/internal/eventbus"
	"multiverse-core/services/realitymonitor"
)

func main() {
	log.Println("Starting Reality Monitor service...")

	// Initialize event bus
	kafkaBrokers := []string{"localhost:9092"} // Default Kafka broker address
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		// Split brokers by comma if multiple are provided
		// For simplicity, we'll use the default for now
	}
	
	eventBus := eventbus.NewEventBus(kafkaBrokers)

	// Create Reality Monitor service
	service := realitymonitor.NewService(eventBus)

	// Start the service
	if err := service.Start(); err != nil {
		log.Fatalf("Failed to start Reality Monitor service: %v", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down Reality Monitor service...")
	
	// Stop the service
	service.Stop()
	
	// Give some time for graceful shutdown
	time.Sleep(1 * time.Second)
	
	log.Println("Reality Monitor service stopped")
}