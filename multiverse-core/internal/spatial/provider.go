// pkg/spatial/provider.go

package spatial

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GeometryProvider интерфейс для получения геометрии.
type GeometryProvider interface {
	GetGeometry(ctx context.Context, worldID, entityID string) (*Geometry, error)
}

// SemanticMemoryProvider реализует получение через HTTP.
type SemanticMemoryProvider struct {
	BaseURL string
	Client  *http.Client
}

func NewSemanticMemoryProvider(baseURL string) *SemanticMemoryProvider {
	return &SemanticMemoryProvider{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (p *SemanticMemoryProvider) GetGeometry(ctx context.Context, worldID, entityID string) (*Geometry, error) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"entity_id": entityID,
		"fields":    []string{"geometry"},
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/v1/entity/"+entityID,
		bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var result struct {
		Geometry Geometry `json:"geometry"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Geometry, nil
}

// StaticProvider для тестов и простых сценариев.
type StaticProvider map[string]*Geometry

func (sp StaticProvider) GetGeometry(_ context.Context, _, entityID string) (*Geometry, error) {
	if g, ok := sp[entityID]; ok {
		return g, nil
	}
	return nil, fmt.Errorf("geometry not found: %s", entityID)
}
