// Package semanticmemory handles ChromaDB integration using native HTTP requests.
package semanticmemory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// ChromaClient handles communication with ChromaDB via HTTP API.
type ChromaClient struct {
	baseURL        string
	httpClient     *http.Client
	collectionID   string // Кэшируем ID коллекции
	collectionName string // Имя коллекции
}

// NewChromaClient creates a new ChromaClient using native HTTP.
func NewChromaClient() *ChromaClient {
	url := os.Getenv("CHROMA_URL")
	if url == "" {
		url = "http://chromadb:8000" // Убедитесь, что это правильный URL
	}
	return &ChromaClient{
		baseURL: url,
		// Настройте HTTP-клиент с таймаутами
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		collectionName: "world_memory", // Фиксированное имя коллекции
	}
}

// --- Внутренние структуры для JSON ---

type chromaCollectionRequest struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata,omitempty"` // Используем map[string]interface{} для метаданных коллекции
}

// Ответ на создание коллекции
type chromaCollectionResponse struct {
	Name     string                 `json:"name"`
	ID       string                 `json:"id"` // Важно: это UUID
	Metadata map[string]interface{} `json:"metadata"`
}

// Структура для элемента списка коллекций
type chromaCollectionItem struct {
	Name     string                 `json:"name"`
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata"`
}

// --- Методы ChromaClient ---

// getOrCreateCollectionID получает ID коллекции по имени, создавая её при необходимости.
// Исправлено: GET /api/v1/collections возвращает массив.
func (c *ChromaClient) getOrCreateCollectionID(ctx context.Context) (string, error) {
	// Проверяем кэш
	if c.collectionID != "" {
		return c.collectionID, nil
	}

	// Сначала получаем список всех коллекций
	endpoint := fmt.Sprintf("%s/api/v1/collections", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request to list collections: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute list collections request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Читаем тело для лога
		return "", fmt.Errorf("list collections request failed with status %d: %s", resp.StatusCode, string(body))
	}

	//log.Println(json.NewDecoder(resp.Body))
	// ИСПРАВЛЕНО: Декодируем в слайс, а не в структуру с полем collections
	var listResult []chromaCollectionItem
	if err := json.NewDecoder(resp.Body).Decode(&listResult); err != nil {
		return "", fmt.Errorf("failed to decode list collections response: %w", err)
	}

	// Ищем коллекцию по имени
	for _, col := range listResult {
		if col.Name == c.collectionName {
			log.Printf("Found existing collection: %s with ID: %s", col.Name, col.ID)
			c.collectionID = col.ID // Кэшируем ID
			return c.collectionID, nil
		}
	}

	// Коллекция не найдена, создаём её
	log.Printf("Collection '%s' not found, creating it...", c.collectionName)
	endpoint = fmt.Sprintf("%s/api/v1/collections", c.baseURL)
	payload := chromaCollectionRequest{
		Name: c.collectionName,
		// Metadata: nil, // nil будет передано как null в JSON, что обычно нормально. Или можно передать пустой map.
		Metadata: map[string]interface{}{}, // Пустой map для ясности, если сервер ожидает объект
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal create collection request: %w", err)
	}

	req, err = http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request for create collection: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute create collection request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated { // Обычно 201 для созданного ресурса
		body, _ := io.ReadAll(resp.Body) // Читаем тело для лога
		return "", fmt.Errorf("create collection request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Декодируем ответ, чтобы получить ID новой коллекции
	var createResult chromaCollectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResult); err != nil {
		return "", fmt.Errorf("failed to decode create collection response: %w", err)
	}

	log.Printf("Created collection: %s with ID: %s", createResult.Name, createResult.ID)
	c.collectionID = createResult.ID // Кэшируем ID
	return c.collectionID, nil
}

// UpsertDocument adds or updates a document in ChromaDB via HTTP POST to /api/v1/collections/{collection_id}/upsert.
func (c *ChromaClient) UpsertDocument(ctx context.Context, entityID string, text string, metadata map[string]interface{}) error {
	// Получаем ID коллекции (создаём, если нужно)
	collectionID, err := c.getOrCreateCollectionID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get/create collection ID for upsert: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/v1/collections/%s/upsert", c.baseURL, collectionID)

	payload := map[string]interface{}{
		"ids":       []string{entityID},
		"documents": []string{text},
		"metadatas": []map[string]interface{}{metadata}, // ChromaDB ожидает массив метаданных
		// "embeddings": nil, // Полагаемся на серверный эмбеддер
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal upsert request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute upsert request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Читаем тело для лога
		return fmt.Errorf("upsert request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// log.Printf("Successfully upserted document for entity: %s", entityID)
	return nil
}

// GetDocuments retrieves documents by their IDs via HTTP POST to /api/v1/collections/{collection_id}/get.
func (c *ChromaClient) GetDocuments(ctx context.Context, entityIDs []string) (map[string]string, error) {
	// Получаем ID коллекции
	collectionID, err := c.getOrCreateCollectionID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection ID for get: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/v1/collections/%s/get", c.baseURL, collectionID)

	// ChromaDB v2 API для Get позволяет фильтровать по ID через тело запроса
	payload := map[string]interface{}{
		"ids": entityIDs,
		// "where": nil, // Дополнительные фильтры
		// "where_document": nil, // Фильтры по документу
		"include": []string{"documents", "metadatas", "embeddings"}, // Что включать в ответ
		// Для простоты, ChromaDB обычно возвращает все по умолчанию или документы, если не указано иное
	}
	log.Println(payload)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal get request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Читаем тело для лога
		return nil, fmt.Errorf("get request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Ids        []string                 `json:"ids"`
		Documents  []string                 `json:"documents"`
		Metadatas  []map[string]interface{} `json:"metadatas"`
		Embeddings [][]float32              `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode get response: %w", err)
	}

	// Create map from IDs to documents
	contexts := make(map[string]string)
	// Ensure lengths match to avoid panic
	if len(result.Ids) != len(result.Documents) {
		return nil, fmt.Errorf("mismatched lengths in ChromaDB get response: ids=%d, docs=%d", len(result.Ids), len(result.Documents))
	}

	for i, id := range result.Ids {
		if i < len(result.Documents) {
			contexts[id] = result.Documents[i]
		}
	}

	return contexts, nil
}

// SearchEventsByType searches for events by type in ChromaDB.
func (c *ChromaClient) SearchEventsByType(ctx context.Context, eventType string, limit int) ([]string, error) {
	// Получаем ID коллекции
	collectionID, err := c.getOrCreateCollectionID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection ID for search: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/v1/collections/%s/query", c.baseURL, collectionID)

	// Создаем фильтр для поиска событий по типу
	whereClause := map[string]interface{}{
		"event_type": eventType,
	}

	payload := map[string]interface{}{
		"where":     whereClause,
		"n_results": limit,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Читаем тело для лога
		return nil, fmt.Errorf("search request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Ids        []string                 `json:"ids"`
		Documents  []string                 `json:"documents"`
		Metadatas  []map[string]interface{} `json:"metadatas"`
		Embeddings [][]float32              `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return result.Documents, nil
}

// Close closes the HTTP client (optional).
func (c *ChromaClient) Close() error {
	// http.Client не требует закрытия, если не использовался custom Transport с закрываемыми ресурсами.
	// c.httpClient.CloseIdleConnections() // Опционально
	return nil
}
