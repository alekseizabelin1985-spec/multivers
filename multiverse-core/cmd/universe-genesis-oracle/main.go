// cmd/universe-genesis-oracle/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"multiverse-core/internal/eventbus"
	"multiverse-core/services/universegenesis"
)

func main() {
	// Инициализация EventBus
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"redpanda:9092"}
	}
	bus := eventbus.NewEventBus(brokers)

	// Инициализация клиента для OntologicalArchivist
	archivistClient := universegenesis.NewArchivistClient(os.Getenv("ARCHIVIST_URL"))

	// Инициализация сервиса
	service := universegenesis.NewService(bus, archivistClient)

	// Обработка сигналов для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down UniverseGenesisOracle...")
		cancel()
	}()

	log.Println("UniverseGenesisOracle starting...")
	if err := service.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal("Service failed:", err)
	}
	log.Println("UniverseGenesisOracle completed its task and stopped.")
}
