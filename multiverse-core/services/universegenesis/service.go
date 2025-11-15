// services/universegenesis/service.go
package universegenesis

import (
	"context"
	"log"

	"multiverse-core/internal/eventbus"
	"multiverse-core/internal/oracle" // <-- Импорт общего клиента

	"github.com/google/uuid"
)

type Service struct {
	bus       *eventbus.EventBus
	archivist *ArchivistClient
	oracle    *oracle.Client // <-- Изменён тип
	generator *Generator
}

func NewService(bus *eventbus.EventBus, archivist *ArchivistClient) *Service { // <-- Принимаем URL, а не готовый клиент
	oracleClient := oracle.NewClient() // <-- Создаём общий клиент
	return &Service{
		bus:       bus,
		archivist: archivist,
		oracle:    oracleClient,                               // <-- Передаём общий клиент
		generator: NewGenerator(bus, archivist, oracleClient), // <-- Передаём общий клиент
	}
}

func (s *Service) Run(ctx context.Context) error {
	// Пример запуска с жёстко закодированным семенем
	seed := uuid.New().String()
	constraints := []string{} // []string{"no_healing", "ascension_through_suffering"}

	log.Printf("Starting Universe Genesis with seed: %s", seed)
	err := s.generator.StartGenesis(ctx, seed, constraints)
	if err != nil {
		log.Printf("Universe Genesis failed: %v", err)
		return err
	}

	log.Println("Universe Genesis completed successfully.")
	// Сервис завершает работу после завершения генезиса
	return nil
}
