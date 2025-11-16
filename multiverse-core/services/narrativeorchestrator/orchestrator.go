// services/narrativeorchestrator/orchestrator.go

package narrativeorchestrator

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"multiverse-core/internal/config"
	"multiverse-core/internal/eventbus"
	"multiverse-core/internal/minio"
	"multiverse-core/internal/spatial"
)

type SemanticMemoryClient struct {
	BaseURL string
}

type GetContextResponse struct {
	Contexts map[string]string `json:"contexts"`
}

type StateSnapshot struct {
	Entities map[string]interface{} `json:"entities"`
	Canon    []string               `json:"canon"`
	LastMood []string               `json:"last_mood,omitempty"`
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

type NarrativeOrchestrator struct {
	gms         map[string]*GMInstance
	mu          sync.RWMutex
	bus         *eventbus.EventBus
	semantic    *SemanticMemoryClient
	minioClient *minio.Client
	configStore *config.Store
	geoProvider spatial.GeometryProvider
}

func NewNarrativeOrchestrator(bus *eventbus.EventBus) *NarrativeOrchestrator {
	semanticURL := os.Getenv("SEMANTIC_MEMORY_URL")
	if semanticURL == "" {
		semanticURL = "http://semantic-memory:8080"
	}

	minioCfg := minio.Config{
		Endpoint:        os.Getenv("MINIO_ENDPOINT"),
		AccessKeyID:     os.Getenv("MINIO_ACCESS_KEY"),
		SecretAccessKey: os.Getenv("MINIO_SECRET_KEY"),
		Region:          "us-east-1",
	}
	minioClient, _ := minio.NewClient(minioCfg)
	geoProvider := spatial.NewSemanticMemoryProvider(semanticURL)
	configStore := config.NewStore(minioClient, "gnue-configs")

	return &NarrativeOrchestrator{
		gms:         make(map[string]*GMInstance),
		bus:         bus,
		semantic:    &SemanticMemoryClient{BaseURL: semanticURL},
		minioClient: minioClient,
		configStore: configStore,
		geoProvider: geoProvider,
	}
}

func extractIDFromScope(scopeID string) string {
	if parts := strings.SplitN(scopeID, ":", 2); len(parts) == 2 {
		return parts[1]
	}
	return scopeID
}

func getDefaultProfile() *config.Profile {
	return &config.Profile{
		TimeWindow: "10m",
		Triggers: struct {
			TimeIntervalMs    int      `yaml:"time_interval_ms,omitempty" json:"time_interval_ms,omitempty"`
			MaxEvents         int      `yaml:"max_events,omitempty" json:"max_events,omitempty"`
			NarrativeTriggers []string `yaml:"narrative_triggers,omitempty" json:"narrative_triggers,omitempty"`
		}{
			TimeIntervalMs: 10000,
			MaxEvents:      50,
		},
	}
}

func (no *NarrativeOrchestrator) CreateGM(ev eventbus.Event) {
	scopeID, _ := ev.Payload["scope_id"].(string)
	scopeType, _ := ev.Payload["scope_type"].(string)
	if scopeType == "" {
		if parts := strings.SplitN(scopeID, ":", 2); len(parts) == 2 {
			scopeType = parts[0]
		} else {
			scopeType = "unknown"
		}
	}

	profile, _ := no.configStore.GetProfile(scopeType)
	if profile == nil {
		profile = getDefaultProfile()
	}

	override, _ := no.configStore.GetOverride(scopeID)
	if override != nil {
		profile = config.MergeProfiles(profile, override)
	}

	focusEntities := make([]string, len(profile.FocusEntities))
	for i, tpl := range profile.FocusEntities {
		if strings.Contains(tpl, "{{.player_id}}") {
			focusEntities[i] = strings.Replace(tpl, "{{.player_id}}", extractIDFromScope(scopeID), -1)
		} else if strings.Contains(tpl, "{{.group_id}}") {
			focusEntities[i] = strings.Replace(tpl, "{{.group_id}}", extractIDFromScope(scopeID), -1)
		} else {
			focusEntities[i] = tpl
		}
	}

	gm := &GMInstance{
		ScopeID:   scopeID,
		ScopeType: scopeType,
		WorldID:   ev.WorldID,
		State:     make(map[string]interface{}),
		History:   []HistoryEntry{},
		Config:    profile.ToMap(),
		CreatedAt: time.Now(),
	}

	geometry, _ := no.geoProvider.GetGeometry(context.Background(), ev.WorldID, scopeID)
	if geometry == nil {
		geometry = &spatial.Geometry{Point: &spatial.Point{X: 0, Y: 0}}
	}

	gm.VisibilityScope = spatial.DefaultScope(scopeType, geometry, gm.Config)
	if scopeType == "location" {
		gm.VisibilityScope = gm.VisibilityScope.Buffer(200)
	}

	if savedGM, _ := no.loadSnapshot(scopeID); savedGM != nil {
		gm = savedGM
		gm.ScopeID = scopeID
		gm.ScopeType = scopeType
		gm.WorldID = ev.WorldID
		gm.Config = profile.ToMap()
		geometry, _ = no.geoProvider.GetGeometry(context.Background(), ev.WorldID, scopeID)
		if geometry != nil {
			gm.VisibilityScope = spatial.DefaultScope(scopeType, geometry, gm.Config)
			if scopeType == "location" {
				gm.VisibilityScope = gm.VisibilityScope.Buffer(200)
			}
		}
	}

	timeoutMin := 30.0
	if cfgTriggers, ok := gm.Config["triggers"].(map[string]interface{}); ok {
		if intervalMs, ok := cfgTriggers["time_interval_ms"].(float64); ok {
			timeoutMin = intervalMs / 60000.0 * 5
		}
	}
	time.AfterFunc(time.Duration(timeoutMin)*time.Minute, func() {
		no.DeleteGMByScope(scopeID)
	})

	no.mu.Lock()
	no.gms[scopeID] = gm
	no.mu.Unlock()

	log.Printf("GM created: %s (type: %s)", scopeID, scopeType)
}

func (no *NarrativeOrchestrator) extractEventPoint(ev eventbus.Event) (spatial.Point, bool) {
	if loc, ok := ev.Payload["location"].(map[string]interface{}); ok {
		x := loc["x"].(float64)
		y := loc["y"].(float64)
		return spatial.Point{X: x, Y: y}, true
	}
	if to, ok := ev.Payload["to"].(map[string]interface{}); ok {
		x := to["x"].(float64)
		y := to["y"].(float64)
		return spatial.Point{X: x, Y: y}, true
	}
	return spatial.Point{}, false
}

func (no *NarrativeOrchestrator) DeleteGMByScope(scopeID string) {
	no.mu.Lock()
	defer no.mu.Unlock()
	delete(no.gms, scopeID)
}

func (no *NarrativeOrchestrator) DeleteGM(ev eventbus.Event) {
	scopeID, _ := ev.Payload["scope_id"].(string)
	no.mu.Lock()
	delete(no.gms, scopeID)
	no.mu.Unlock()
	log.Printf("GM deleted: %s", scopeID)
}

func (no *NarrativeOrchestrator) MergeGM(ev eventbus.Event) {}
func (no *NarrativeOrchestrator) SplitGM(ev eventbus.Event) {}

func (no *NarrativeOrchestrator) loadSnapshot(scopeID string) (*GMInstance, error) {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(scopeID)))
	prefix := path.Join("gnue", "gm-snapshots", "v1", hash)
	objects, _ := no.minioClient.ListObjects("gnue-snapshots", prefix)
	if len(objects) == 0 {
		return nil, fmt.Errorf("no snapshots")
	}
	data, _ := no.minioClient.GetObject("gnue-snapshots", objects[0].Key)
	var gm GMInstance
	json.Unmarshal(data, &gm)
	return &gm, nil
}

func (no *NarrativeOrchestrator) saveSnapshot(scopeID string, gm *GMInstance) error {
	data, _ := json.Marshal(gm)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(scopeID)))
	timestamp := time.Now().Unix()
	key := path.Join("gnue", "gm-snapshots", "v1", hash, fmt.Sprintf("%d_001.json", timestamp))
	return no.minioClient.PutObject("gnue-snapshots", key, bytes.NewReader(data), int64(len(data)))
}

func (no *NarrativeOrchestrator) HandleGameEvent(ev eventbus.Event) {
	eventPoint, ok := no.extractEventPoint(ev)
	if !ok {
		return
	}

	no.mu.RLock()
	gmsToNotify := []*GMInstance{}
	for _, gm := range no.gms {
		if gm.VisibilityScope.IsInScope(eventPoint) {
			gmsToNotify = append(gmsToNotify, gm)
		}
	}
	no.mu.RUnlock()

	for _, gm := range gmsToNotify {
		no.processEventForGM(ev, gm)
	}
}

func (no *NarrativeOrchestrator) processEventForGM(ev eventbus.Event, gm *GMInstance) {
	no.mu.Lock()
	gm.History = append(gm.History, HistoryEntry{
		EventID:   ev.EventID,
		Timestamp: ev.Timestamp,
	})
	no.mu.Unlock()

	entityIDs := extractEntityIDs(ev.Payload)
	entityIDs = append(entityIDs, gm.ScopeID)

	contexts, _ := no.semantic.GetContext(context.Background(), entityIDs)
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

	systemPrompt, userPrompt := BuildPrompts(worldContext, gm.ScopeID, gm.ScopeType, entitiesContext, triggerEvent)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	oracleResp, _ := CallOracle(ctx, systemPrompt, userPrompt)

	if len(oracleResp.Mood) > 0 {
		no.mu.Lock()
		if gm.State == nil {
			gm.State = make(map[string]interface{})
		}
		gm.State["last_mood"] = oracleResp.Mood
		no.mu.Unlock()
	}

	outputEvent := eventbus.Event{
		EventID:   "narrative-" + time.Now().Format("20060102-150405-") + ev.EventID[:8],
		EventType: "narrative.generated",
		Source:    "narrative-orchestrator",
		WorldID:   ev.WorldID,
		ScopeID:   &gm.ScopeID,
		Payload: map[string]interface{}{
			"narrative":  oracleResp.Narrative,
			"mood":       oracleResp.Mood,
			"new_events": oracleResp.NewEvents,
			"trigger":    ev.EventID,
			"gm_scope":   gm.ScopeID,
		},
		Timestamp: time.Now(),
	}

	no.bus.Publish(context.Background(), eventbus.TopicNarrativeOutput, outputEvent)
	no.saveSnapshot(gm.ScopeID, gm)
	log.Printf("GM %s generated narrative", gm.ScopeID)
}

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
