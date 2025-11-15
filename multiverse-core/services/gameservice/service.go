package gameservice

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"multiverse-core/internal/entity"
	"multiverse-core/internal/eventbus"
	"time"
)

type Config struct {
	KafkaBrokers []string
	HTTPAddr     string
	CacheTTL     time.Duration
}

type Service struct {
	bus           *eventbus.EventBus
	httpServer    *HTTPServer
	wsServer      *WebSocketServer
	entityCache   *EntityCache
	minioClient   *MinioClient
	playerService *PlayerService
	broadcast     chan []byte
	cfg           Config
}

func NewService(cfg Config) *Service {
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = time.Minute * 5 // По умолчанию 5 минут
	}

	bus := eventbus.NewEventBus(cfg.KafkaBrokers)
	minioClient, err := NewMinioClient()
	if err != nil {
		log.Printf("Warning: Failed to create MinIO client: %v", err)
	}

	playerService := NewPlayerService(NewEntityCache(cfg.CacheTTL), minioClient, bus)

	return &Service{
		bus:           bus,
		httpServer:    NewHTTPServer(cfg.HTTPAddr),
		wsServer:      NewWebSocketServer(),
		entityCache:   NewEntityCache(cfg.CacheTTL),
		minioClient:   minioClient,
		playerService: playerService,
		broadcast:     make(chan []byte, 100), // Буферизованный канал для broadcast сообщений
		cfg:           cfg,
	}
}

func (s *Service) Start(ctx context.Context) {
	log.Println("GameService started. Initializing components...")

	// Загружаем начальные данные из MinIO
	s.loadInitialData(ctx)

	// Запуск HTTP сервера
	s.httpServer.RegisterRoutes(s, s.wsServer)
	s.httpServer.Start()

	// Запуск WebSocket сервера
	go s.wsServer.BroadcastLoop(s.broadcast)

	// Подписываемся на топики событий
	topics := []string{
		eventbus.TopicWorldEvents,
		eventbus.TopicGameEvents,
		eventbus.TopicPlayerEvents,
		eventbus.TopicSystemEvents,
		eventbus.TopicNarrativeOutput,
	}

	for _, topic := range topics {
		topic := topic
		go func() {
			s.bus.Subscribe(ctx, topic, "game-service-group", s.handleEvent)
		}()
	}

	// Запуск обработчиков событий
	entityHandler := NewEntityStreamHandler(s.entityCache, s.broadcast, s.minioClient)
	eventHandler := NewEventStreamHandler(s.broadcast)

	// Передаем обработчики в основной цикл
	go s.eventProcessingLoop(entityHandler, eventHandler)

	log.Println("GameService fully initialized and running.")
}

func (s *Service) Stop() {
	s.bus.Close()
	s.httpServer.Stop()
	close(s.broadcast)
}

func (s *Service) handleEvent(event eventbus.Event) {
	// Определяем тип события и передаем его соответствующему обработчику
	switch {
	case len(event.EventType) >= len(eventbus.TypeEntity) && event.EventType[:len(eventbus.TypeEntity)] == eventbus.TypeEntity:
		// События сущностей
		entityHandler := NewEntityStreamHandler(s.entityCache, s.broadcast, s.minioClient)
		entityHandler.HandleEntityEvent(event)
	case len(event.EventType) >= len(eventbus.TypeNarrative) && event.EventType[:len(eventbus.TypeNarrative)] == eventbus.TypeNarrative:
		// Повествовательные события
		// TODO: Обработать повествовательные события
		// message, _ := json.Marshal(map[string]interface{}{
		// 	"type":  "narrative_event",
		// 	"event": event,
		// })
		message, _ := json.Marshal(event)
		s.broadcast <- message
	default:
		// Остальные игровые события
		eventHandler := NewEventStreamHandler(s.broadcast)
		eventHandler.HandleGameEvent(event)
	}
}

func (s *Service) eventProcessingLoop(entityHandler *EntityStreamHandler, eventHandler *EventStreamHandler) {
	for message := range s.broadcast {
		// Отправляем сообщение всем подключенным WebSocket клиентам
		s.wsServer.BroadcastMessage(message)
	}
}

func (s *Service) loadInitialData(ctx context.Context) {
	if s.minioClient == nil {
		log.Println("MinIO client not available, skipping initial data load")
		return
	}

	// TODO: Загрузить начальные данные из MinIO
	// Это может включать:
	// 1. Загрузку часто используемых сущностей в кэш
	// 2. Загрузку последних событий
	// 3. Инициализацию игрового состояния

	log.Println("Initial data loading completed")
}

// GetEntity получает сущность из кэша или MinIO
func (s *Service) GetEntity(ctx context.Context, entityID, worldID string) (*entity.Entity, error) {
	// Пытаемся получить сущность из кэша
	if entity, found := s.entityCache.Get(entityID, worldID); found {
		return entity, nil
	}

	// Если сущности нет в кэше и доступен MinIO клиент, загружаем из MinIO
	if s.minioClient != nil {
		entity, err := s.minioClient.LoadEntity(ctx, entityID, worldID)
		if err != nil {
			return nil, err
		}

		// Сохраняем в кэш
		s.entityCache.Set(entityID, worldID, entity)
		return entity, nil
	}

	return nil, fmt.Errorf("entity not found in cache and MinIO client not available")
}

// RegisterPlayer регистрирует нового игрока
func (s *Service) RegisterPlayer(ctx context.Context, playerID, playerName, worldID string) (*entity.Entity, error) {
	return s.playerService.RegisterPlayer(ctx, playerID, playerName, worldID)
}

// LoginPlayer выполняет вход игрока
func (s *Service) LoginPlayer(ctx context.Context, playerID, worldID string) (*entity.Entity, error) {
	return s.playerService.LoginPlayer(ctx, playerID, worldID)
}

// GetOrCreatePlayer получает существующего игрока или создает нового
func (s *Service) GetOrCreatePlayer(ctx context.Context, playerID, playerName, worldID string) (*entity.Entity, error) {
	return s.playerService.GetOrCreatePlayer(ctx, playerID, playerName, worldID)
}
