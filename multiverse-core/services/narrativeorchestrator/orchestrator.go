// services/narrativeorchestrator/orchestrator.go
package narrativeorchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"multiverse-core/internal/eventbus"
)

// SemanticMemoryClient fetches context from Semantic Memory Builder.
type SemanticMemoryClient struct {
	BaseURL string
}

type GetContextResponse struct {
	Contexts map[string]string `json:"contexts"`
}

func (c *SemanticMemoryClient) GetContext(ctx context.Context, entityIDs []string) (map[string]string, error) {
	if len(entityIDs) == 0 {
		return map[string]string{}, nil
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"entity_ids": entityIDs,
		"depth":      2,
	})

	resp, err := http.Post(c.BaseURL+"/v1/context", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result GetContextResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Contexts, nil
}

// NarrativeOrchestrator manages GM instances.
type NarrativeOrchestrator struct {
	gms      map[string]*GMInstance
	mu       sync.RWMutex
	bus      *eventbus.EventBus
	semantic *SemanticMemoryClient
}

// NewNarrativeOrchestrator creates a new orchestrator.
func NewNarrativeOrchestrator(bus *eventbus.EventBus) *NarrativeOrchestrator {
	semanticURL := os.Getenv("SEMANTIC_MEMORY_URL")
	if semanticURL == "" {
		semanticURL = "http://semantic-memory:8080"
	}

	return &NarrativeOrchestrator{
		gms:      make(map[string]*GMInstance),
		bus:      bus,
		semantic: &SemanticMemoryClient{BaseURL: semanticURL},
	}
}

// CreateGM handles gm.created event.
func (no *NarrativeOrchestrator) CreateGM(ev eventbus.Event) {
	scopeID, ok := ev.Payload["scope_id"].(string)
	if !ok {
		log.Printf("[event=%s] Invalid gm.created: missing scope_id", ev.EventID)
		return
	}

	scopeType, _ := ev.Payload["scope_type"].(string)
	config, _ := ev.Payload["config"].(map[string]interface{})
	if config == nil {
		config = map[string]interface{}{"timeout_minutes": 30.0}
	}

	gm := &GMInstance{
		ScopeID:   scopeID,
		ScopeType: scopeType,
		WorldID:   ev.WorldID,
		State:     make(map[string]interface{}),
		History:   []string{},
		Config:    config,
		CreatedAt: time.Now(),
	}

	// Set timeout
	if timeoutMin, ok := config["timeout_minutes"].(float64); ok && timeoutMin > 0 {
		time.AfterFunc(time.Duration(timeoutMin)*time.Minute, func() {
			no.DeleteGMByScope(scopeID)
		})
	}

	no.mu.Lock()
	no.gms[scopeID] = gm
	no.mu.Unlock()

	log.Printf("[event=%s] GM created for scope %s (type: %s)", ev.EventID, scopeID, scopeType)
}

// DeleteGMByScope safely deletes a GM by scope_id.
func (no *NarrativeOrchestrator) DeleteGMByScope(scopeID string) {
	no.mu.Lock()
	defer no.mu.Unlock()
	if _, exists := no.gms[scopeID]; exists {
		delete(no.gms, scopeID)
		log.Printf("GM auto-deleted for inactive scope %s", scopeID)
	}
}

// DeleteGM handles gm.deleted event.
func (no *NarrativeOrchestrator) DeleteGM(ev eventbus.Event) {
	scopeID, ok := ev.Payload["scope_id"].(string)
	if !ok {
		log.Printf("[event=%s] Invalid gm.deleted: missing scope_id", ev.EventID)
		return
	}

	no.mu.Lock()
	delete(no.gms, scopeID)
	no.mu.Unlock()

	log.Printf("[event=%s] GM deleted for scope %s", ev.EventID, scopeID)
}

// MergeGM handles gm.merged event (stub).
func (no *NarrativeOrchestrator) MergeGM(ev eventbus.Event) {
	log.Printf("[event=%s] gm.merged received (stub): %v", ev.EventID, ev.Payload)
	// TODO: Implement merging logic
}

// SplitGM handles gm.split event (stub).
func (no *NarrativeOrchestrator) SplitGM(ev eventbus.Event) {
	log.Printf("[event=%s] gm.split received (stub): %v", ev.EventID, ev.Payload)
	// TODO: Implement splitting logic
}

// HandleGameEvent processes world_events and game_events for active GMs.
func (no *NarrativeOrchestrator) HandleGameEvent(ev eventbus.Event) {
	if ev.ScopeID == nil {
		return
	}
	scopeID := *ev.ScopeID

	// Блокируем на чтение
	no.mu.RLock()
	gm, exists := no.gms[scopeID]
	if !exists {
		no.mu.RUnlock()
		return
	}
	// Копируем необходимые данные для работы без блокировки
	gmCopy := *gm
	no.mu.RUnlock()

	// Обновляем историю (требует записи)
	no.mu.Lock()
	if currentGM, exists := no.gms[scopeID]; exists {
		currentGM.History = append(currentGM.History, ev.EventID)
	}
	no.mu.Unlock()

	// Извлекаем entity IDs
	entityIDs := extractEntityIDs(ev.Payload)
	entityIDs = append(entityIDs, scopeID)

	// Получаем контекст
	contexts, err := no.semantic.GetContext(context.Background(), entityIDs)
	if err != nil {
		log.Printf("[event=%s] Failed to get context for GM %s: %v", ev.EventID, scopeID, err)
		return
	}

	worldContext := contexts[ev.WorldID]
	if worldContext == "" {
		worldContext = "Нет данных о мире"
	}

	var entitiesLines []string
	for _, id := range entityIDs {
		if ctx, exists := contexts[id]; exists {
			entitiesLines = append(entitiesLines, ctx)
		}
	}
	entitiesContext := "Нет данных"
	if len(entitiesLines) > 0 {
		entitiesContext = strings.Join(entitiesLines, "\n")
	}

	triggerEvent := "Неизвестное событие"
	if desc, exists := ev.Payload["description"].(string); exists {
		triggerEvent = desc
	} else if mentions, exists := ev.Payload["mentions"].([]interface{}); exists {
		triggerEvent = "Событие с участием: " + strings.Join(toStringSlice(mentions), ", ")
	}

	// Генерация повествования
	prompt := BuildPrompt(worldContext, scopeID, gmCopy.ScopeType, entitiesContext, triggerEvent)

	oracleResp, err := CallOracle(context.Background(), prompt)
	if err != nil {
		log.Printf("[event=%s] Oracle call failed for GM %s: %v", ev.EventID, scopeID, err)
		return
	}

	// Публикация результата
	outputEvent := eventbus.Event{
		EventID:   "narrative-" + time.Now().Format("20060102-150405-") + ev.EventID[:8],
		EventType: "narrative.generated",
		Source:    "narrative-orchestrator",
		WorldID:   ev.WorldID,
		ScopeID:   &scopeID,
		Payload: map[string]interface{}{
			"narrative":  oracleResp.Narrative,
			"new_events": oracleResp.NewEvents,
			"trigger":    ev.EventID,
			"gm_scope":   scopeID,
		},
		Timestamp: time.Now(),
	}

	no.bus.Publish(context.Background(), eventbus.TopicNarrativeOutput, outputEvent)
	log.Printf("[event=%s] GM %s generated narrative", ev.EventID, scopeID)
}

// Helper functions
func extractEntityIDs(payload map[string]interface{}) []string {
	var ids []string

	if mentions, ok := payload["mentions"].([]interface{}); ok {
		for _, m := range mentions {
			if id, ok := m.(string); ok {
				ids = append(ids, id)
			}
		}
	}

	if entityID, ok := payload["entity_id"].(string); ok {
		ids = append(ids, entityID)
	}
	if target, ok := payload["target"].(string); ok {
		ids = append(ids, target)
	}
	if source, ok := payload["source"].(string); ok {
		ids = append(ids, source)
	}

	return ids
}

func toStringSlice(v []interface{}) []string {
	res := make([]string, len(v))
	for i, val := range v {
		if s, ok := val.(string); ok {
			res[i] = s
		}
	}
	return res
}
