// Package main is the entry point for OntologicalArchivist.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"multiverse-core/services/ontologicalarchivist"

	"github.com/gorilla/mux"
)

func main() {
	// Create service
	cfg := ontologicalarchivist.Config{
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "minio:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		KafkaBrokers:   getEnvBrokers("KAFKA_BROKERS", []string{"redpanda:9092"}),
	}
	service := ontologicalarchivist.NewService(cfg)

	// Setup HTTP server
	r := mux.NewRouter()
	service.SetupRoutes(r)

	ONTOLOGICAL_PORT := os.Getenv("ONTOLOGICAL_PORT")
	if ONTOLOGICAL_PORT == "" {
		ONTOLOGICAL_PORT = "8081"
	}

	server := &http.Server{
		Addr:         ":" + ONTOLOGICAL_PORT,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Handle shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down OntologicalArchivist...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)

		cancel()
	}()

	log.Println("OntologicalArchivist starting on :" + ONTOLOGICAL_PORT)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed:", err)
	}
	log.Println("OntologicalArchivist stopped.")
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
