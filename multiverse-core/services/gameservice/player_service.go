package gameservice

import (
	"context"
	"fmt"
	"log"
	"multiverse-core/internal/entity"
	"multiverse-core/internal/eventbus"
	"time"
)

// PlayerService управляет регистрацией и входом игроков
type PlayerService struct {
	entityCache *EntityCache
	minioClient *MinioClient
	eventBus    *eventbus.EventBus
}

// NewPlayerService создает новый сервис управления игроками
func NewPlayerService(entityCache *EntityCache, minioClient *MinioClient, eventBus *eventbus.EventBus) *PlayerService {
	return &PlayerService{
		entityCache: entityCache,
		minioClient: minioClient,
		eventBus:    eventBus,
	}
}

// RegisterPlayer регистрирует нового игрока, создавая сущность если её нет
func (ps *PlayerService) RegisterPlayer(ctx context.Context, playerID, playerName, worldID string) (*entity.Entity, error) {
	// Пытаемся получить сущность игрока из кэша
	if entity, found := ps.entityCache.Get(playerID, worldID); found {
		return entity, nil
	}

	// Если сущности нет в кэше, пытаемся загрузить из MinIO
	playerEntity, err := ps.minioClient.LoadEntity(ctx, playerID, worldID)
	if err == nil {
		// Сущность найдена в MinIO, добавляем в кэш
		ps.entityCache.Set(playerID, worldID, playerEntity)
		return playerEntity, nil
	}

	// Если сущность не найдена, создаем новую
	playerEntity = entity.NewEntity(playerID, "player", map[string]interface{}{
		"name": playerName,
		"location": map[string]interface{}{
			"world_id": worldID,
		},
	})

	// Сохраняем новую сущность в MinIO
	err = ps.minioClient.SaveEntity(ctx, playerEntity, worldID)
	if err != nil {
		return nil, fmt.Errorf("failed to save new player entity: %w", err)
	}

	// Добавляем в кэш
	ps.entityCache.Set(playerID, worldID, playerEntity)

	// Публикуем событие о создании игрока
	event := eventbus.Event{
		EventID:   fmt.Sprintf("player_created_%d", time.Now().Unix()),
		EventType: "entity.created",
		WorldID:   worldID,
		Timestamp: time.Now().UTC(),
		Payload: map[string]interface{}{
			"entity_id":   playerID,
			"entity_type": "player",
			"payload":     playerEntity.Payload,
		},
	}

	err = ps.eventBus.Publish(ctx, eventbus.TopicSystemEvents, event)
	if err != nil {
		log.Printf("Warning: Failed to publish player created event: %v", err)
	}

	return playerEntity, nil
}

// LoginPlayer выполняет вход игрока (возвращает существующую сущность)
func (ps *PlayerService) LoginPlayer(ctx context.Context, playerID, worldID string) (*entity.Entity, error) {
	// Пытаемся получить сущность игрока из кэша
	if entity, found := ps.entityCache.Get(playerID, worldID); found {
		return entity, nil
	}

	// Если сущности нет в кэше, пытаемся загрузить из MinIO
	playerEntity, err := ps.minioClient.LoadEntity(ctx, playerID, worldID)
	if err != nil {
		return nil, fmt.Errorf("player not found: %w", err)
	}

	// Добавляем в кэш
	ps.entityCache.Set(playerID, worldID, playerEntity)

	return playerEntity, nil
}

// UpdatePlayerLocation обновляет местоположение игрока
func (ps *PlayerService) UpdatePlayerLocation(ctx context.Context, playerID, worldID, newWorldID string) error {
	// Получаем сущность игрока
	playerEntity, err := ps.LoginPlayer(ctx, playerID, worldID)
	if err != nil {
		return err
	}

	// Обновляем местоположение
	playerEntity.SetPath("location.world_id", newWorldID)

	// Сохраняем обновленную сущность
	err = ps.minioClient.SaveEntity(ctx, playerEntity, worldID)
	if err != nil {
		return fmt.Errorf("failed to save updated player entity: %w", err)
	}

	// Обновляем кэш
	ps.entityCache.Set(playerID, worldID, playerEntity)

	// Публикуем событие о перемещении игрока
	event := eventbus.Event{
		EventID:   fmt.Sprintf("player_moved_%d", time.Now().Unix()),
		EventType: "player.moved",
		WorldID:   worldID,
		Timestamp: time.Now().UTC(),
		Payload: map[string]interface{}{
			"player_id": playerID,
			"from_world": worldID,
			"to_world":   newWorldID,
		},
	}

	err = ps.eventBus.Publish(ctx, eventbus.TopicPlayerEvents, event)
	if err != nil {
		log.Printf("Warning: Failed to publish player moved event: %v", err)
	}

	return nil
}

// GetOrCreatePlayer получает существующего игрока или создает нового
func (ps *PlayerService) GetOrCreatePlayer(ctx context.Context, playerID, playerName, worldID string) (*entity.Entity, error) {
	// Сначала пытаемся найти игрока
	playerEntity, err := ps.LoginPlayer(ctx, playerID, worldID)
	if err == nil {
		return playerEntity, nil
	}

	// Если не найден, создаем нового
	return ps.RegisterPlayer(ctx, playerID, playerName, worldID)
}