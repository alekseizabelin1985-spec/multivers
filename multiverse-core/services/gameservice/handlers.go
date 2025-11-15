package gameservice

import (
	"encoding/json"
	"fmt"
	"log"
	"multiverse-core/internal/eventbus"
	"net/http"

	"github.com/gorilla/mux"
)

// EntityStreamHandler обрабатывает поток обновлений сущностей
type EntityStreamHandler struct {
	entityCache *EntityCache
	broadcast   chan []byte
	minioClient *MinioClient
}

func NewEntityStreamHandler(entityCache *EntityCache, broadcast chan []byte, minioClient *MinioClient) *EntityStreamHandler {
	return &EntityStreamHandler{
		entityCache: entityCache,
		broadcast:   broadcast,
		minioClient: minioClient,
	}
}

func (h *EntityStreamHandler) HandleEntityEvent(event eventbus.Event) {
	// Обработка событий, связанных с сущностями
	log.Printf("Handling entity event: %s", event.EventType)

	// TODO: Реализовать логику обработки событий сущностей
	// Это может включать:
	// 1. Получение обновленной сущности из EntityManager (если необходимо)
	// 2. Обновление кэша сущностей
	// 3. Подготовку сообщения для отправки клиентам
	// 4. Отправку сообщения через broadcast канал

	// Пример отправки сообщения клиентам
	// message, _ := json.Marshal(map[string]interface{}{
	// 	"type":  "entity_update",
	// 	"event": event,
	// })
	message, _ := json.Marshal(event)
	h.broadcast <- message
}

// EventStreamHandler обрабатывает поток игровых событий
type EventStreamHandler struct {
	eventBuffer []eventbus.Event
	broadcast   chan []byte
}

func NewEventStreamHandler(broadcast chan []byte) *EventStreamHandler {
	return &EventStreamHandler{
		eventBuffer: make([]eventbus.Event, 0),
		broadcast:   broadcast,
	}
}

func (h *EventStreamHandler) HandleGameEvent(event eventbus.Event) {
	// Обработка игровых событий
	log.Printf("Handling game event: %s", event.EventType)

	// Добавляем событие в буфер
	h.eventBuffer = append(h.eventBuffer, event)

	// Ограничиваем размер буфера (например, 100 последних событий)
	if len(h.eventBuffer) > 100 {
		h.eventBuffer = h.eventBuffer[1:]
	}

	// Подготавливаем сообщение для отправки клиентам
	// message, _ := json.Marshal(map[string]interface{}{
	// 	"type":  "game_event",
	// 	"event": event,
	// })
	message, _ := json.Marshal(event)
	h.broadcast <- message
}

// REST Handlers
func (s *Service) GetEntityHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityID := vars["entity_id"]

	// Получаем worldID из query параметров или заголовков
	worldID := r.URL.Query().Get("world_id")
	if worldID == "" {
		worldID = r.Header.Get("X-World-ID")
	}

	// Если worldID не указан, возвращаем ошибку
	if worldID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("world_id is required"))
		return
	}

	// Пытаемся получить сущность из кэша
	if entity, found := s.entityCache.Get(entityID, worldID); found {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entity)
		return
	}

	// Если сущности нет в кэше и доступен MinIO клиент, загружаем из MinIO
	if s.minioClient != nil {
		entity, err := s.minioClient.LoadEntity(r.Context(), entityID, worldID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Failed to load entity: %v", err)))
			return
		}

		// Сохраняем в кэш
		s.entityCache.Set(entityID, worldID, entity)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entity)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Entity not found"))
}

func (s *Service) RegisterPlayerHandler(w http.ResponseWriter, r *http.Request) {
	// Парсим JSON тело запроса
	var req struct {
		PlayerID   string `json:"player_id"`
		PlayerName string `json:"player_name"`
		WorldID    string `json:"world_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request body"))
		return
	}

	// Проверяем обязательные поля
	if req.PlayerID == "" || req.PlayerName == "" || req.WorldID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("player_id, player_name and world_id are required"))
		return
	}

	// Регистрируем игрока
	entity, err := s.RegisterPlayer(r.Context(), req.PlayerID, req.PlayerName, req.WorldID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to register player: %v", err)))
		return
	}

	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Player registered successfully",
		"player":  entity,
	})
}

func (s *Service) LoginPlayerHandler(w http.ResponseWriter, r *http.Request) {
	// Парсим JSON тело запроса
	var req struct {
		PlayerID string `json:"player_id"`
		WorldID  string `json:"world_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid request body"))
		return
	}

	// Проверяем обязательные поля
	if req.PlayerID == "" || req.WorldID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("player_id and world_id are required"))
		return
	}

	// Входим как игрок
	entity, err := s.LoginPlayer(r.Context(), req.PlayerID, req.WorldID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("Failed to login player: %v", err)))
		return
	}

	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Player logged in successfully",
		"player":  entity,
	})
}

func (s *Service) GetEntityHistoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityID := vars["entity_id"]

	// TODO: Реализовать получение истории сущности по ID
	// Это может включать запрос к EntityManager или другому сервису
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entity_id": entityID,
		"history":   []interface{}{}, // Заглушка
	})
}

func (s *Service) GetRecentEventsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать получение последних событий
	// Это может включать запрос к Event Bus или другому сервису
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": []interface{}{}, // Заглушка
	})
}
