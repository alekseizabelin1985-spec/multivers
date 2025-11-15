// services/universegenesis/archivist.go
package universegenesis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ArchivistClient struct {
	BaseURL string
	Client  *http.Client
}

func NewArchivistClient(baseURL string) *ArchivistClient {
	return &ArchivistClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// SaveSchema теперь использует POST, как и WorldGenerator.
func (ac *ArchivistClient) SaveSchema(ctx context.Context, schemaType, name, version string, schemaData []byte) error {
	// Подготовим тело запроса в формате, ожидаемом OntologicalArchivist
	requestBody, err := json.Marshal(map[string]interface{}{
		"schema_type": schemaType,
		"name":        name,
		"version":     version,
		"schema":      json.RawMessage(schemaData), // Вложенный JSON как RawMessage
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Используем POST к /v1/schemas, как в примерах для OntologicalArchivist
	url := fmt.Sprintf("%s/v1/schemas", ac.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.Client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем, что статус 2xx
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("archivist returned non-2xx status %d: %s", resp.StatusCode, string(body))
	}

	// Логируем успешное сохранение с кодом статуса
	fmt.Printf("DEBUG: Archivist returned status %d for saving schema %s/%s/%s\n", resp.StatusCode, schemaType, name, version)
	return nil

}
