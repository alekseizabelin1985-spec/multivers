// services/entitymanager/service.go
package entitymanager

import (
	"context"
	"log"
	"multiverse-core/internal/eventbus"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	KafkaBrokers   []string
}

type Service struct {
	manager *Manager
	bus     *eventbus.EventBus
}

func NewService(cfg Config) (*Service, error) {
	// Создаём MinIO клиент напрямую
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	manager := &Manager{minio: minioClient}
	bus := eventbus.NewEventBus(cfg.KafkaBrokers)

	return &Service{
		manager: manager,
		bus:     bus,
	}, nil
}

func (s *Service) Start(ctx context.Context) {
	log.Println("EntityManager started (lazy mode). Listening to all event topics...")

	topics := []string{
		eventbus.TopicPlayerEvents,
		eventbus.TopicWorldEvents,
		eventbus.TopicGameEvents,
		eventbus.TopicSystemEvents,
	}

	for _, topic := range topics {
		topic := topic
		go func() {
			s.bus.Subscribe(ctx, topic, "entity-manager-group", s.manager.HandleEvent)
		}()
	}
}

func (s *Service) Stop() {
	s.bus.Close()
}
