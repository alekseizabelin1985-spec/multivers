// internal/config/store.go

package config

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"sync"
	"time"

	"multiverse-core/internal/minio"

	"gopkg.in/yaml.v3"
)

// Profile — структура профиля GM (полностью соответствует YAML).
type Profile struct {
	ScopeType     string   `yaml:"scope_type,omitempty" json:"scope_type,omitempty"`
	FocusEntities []string `yaml:"focus_entities,omitempty" json:"focus_entities,omitempty"`
	TimeWindow    string   `yaml:"time_window,omitempty" json:"time_window,omitempty"`
	ContextDepth  struct {
		Canon    int `yaml:"canon,omitempty" json:"canon,omitempty"`
		History  int `yaml:"history,omitempty" json:"history,omitempty"`
		Entities int `yaml:"entities,omitempty" json:"entities,omitempty"`
	} `yaml:"context_depth,omitempty" json:"context_depth,omitempty"`
	Include struct {
		WorldFacts      bool `yaml:"world_facts,omitempty" json:"world_facts,omitempty"`
		EntityEmotions  bool `yaml:"entity_emotions,omitempty" json:"entity_emotions,omitempty"`
		LocationDetails bool `yaml:"location_details,omitempty" json:"location_details,omitempty"`
		TemporalContext bool `yaml:"temporal_context,omitempty" json:"temporal_context,omitempty"`
	} `yaml:"include,omitempty" json:"include,omitempty"`
	Triggers struct {
		TimeIntervalMs    int      `yaml:"time_interval_ms,omitempty" json:"time_interval_ms,omitempty"`
		MaxEvents         int      `yaml:"max_events,omitempty" json:"max_events,omitempty"`
		NarrativeTriggers []string `yaml:"narrative_triggers,omitempty" json:"narrative_triggers,omitempty"`
	} `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	Snapshot struct {
		IntervalEvents int    `yaml:"interval_events,omitempty" json:"interval_events,omitempty"`
		IntervalMs     int    `yaml:"interval_ms,omitempty" json:"interval_ms,omitempty"`
		MinioPath      string `yaml:"minio_path,omitempty" json:"minio_path,omitempty"`
	} `yaml:"snapshot,omitempty" json:"snapshot,omitempty"`
}

// Store управляет динамическими конфигами в MinIO.
type Store struct {
	minioClient *minio.Client
	cache       map[string]*Profile
	cacheLock   sync.RWMutex
	bucket      string
}

// NewStore создаёт новый config store.
func NewStore(minioClient *minio.Client, bucket string) *Store {
	store := &Store{
		minioClient: minioClient,
		cache:       make(map[string]*Profile),
		bucket:      bucket,
	}
	go store.backgroundRefresh()
	return store
}

// GetProfile возвращает профиль по scope_type (с кэшированием).
func (s *Store) GetProfile(scopeType string) (*Profile, error) {
	s.cacheLock.RLock()
	if p, ok := s.cache[scopeType]; ok {
		s.cacheLock.RUnlock()
		return p, nil
	}
	s.cacheLock.RUnlock()

	key := path.Join("gm-profiles", scopeType+".yaml")
	data, err := s.minioClient.GetObject(s.bucket, key)
	if err != nil {
		return nil, fmt.Errorf("profile %s not found: %w", scopeType, err)
	}

	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("invalid YAML for %s: %w", scopeType, err)
	}

	s.cacheLock.Lock()
	s.cache[scopeType] = &profile
	s.cacheLock.Unlock()

	return &profile, nil
}

// GetOverride возвращает переопределение для scopeID (если есть).
func (s *Store) GetOverride(scopeID string) (*Profile, error) {
	key := path.Join("gm-overrides", scopeID+".yaml")
	data, err := s.minioClient.GetObject(s.bucket, key)
	if err != nil {
		return nil, nil // no override
	}

	var profile Profile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("invalid override for %s: %w", scopeID, err)
	}
	return &profile, nil
}

// backgroundRefresh обновляет кэш каждые 30 сек (hot-reload).
func (s *Store) backgroundRefresh() {
	ticker := time.NewTicker(120 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.cacheLock.Lock()
		s.cache = make(map[string]*Profile)
		s.cacheLock.Unlock()
		log.Println("GM config cache refreshed")
	}
}

// toMap конвертирует Profile → map[string]interface{} для GMInstance.Config.
func (p *Profile) ToMap() map[string]interface{} {
	data, _ := json.Marshal(p)
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	return m
}

// mergeProfiles объединяет базовый профиль и переопределение.
func mergeProfiles(base, override *Profile) *Profile {
	if override == nil {
		return base
	}
	result := *base

	if override.TimeWindow != "" {
		result.TimeWindow = override.TimeWindow
	}
	if override.ScopeType != "" {
		result.ScopeType = override.ScopeType
	}
	if len(override.FocusEntities) > 0 {
		result.FocusEntities = override.FocusEntities
	}
	if override.ContextDepth.Canon != 0 {
		result.ContextDepth.Canon = override.ContextDepth.Canon
	}
	if override.ContextDepth.History != 0 {
		result.ContextDepth.History = override.ContextDepth.History
	}
	if override.ContextDepth.Entities != 0 {
		result.ContextDepth.Entities = override.ContextDepth.Entities
	}
	result.Include.WorldFacts = override.Include.WorldFacts || result.Include.WorldFacts
	result.Include.EntityEmotions = override.Include.EntityEmotions || result.Include.EntityEmotions
	result.Include.LocationDetails = override.Include.LocationDetails || result.Include.LocationDetails
	result.Include.TemporalContext = override.Include.TemporalContext || result.Include.TemporalContext

	if override.Triggers.TimeIntervalMs != 0 {
		result.Triggers.TimeIntervalMs = override.Triggers.TimeIntervalMs
	}
	if override.Triggers.MaxEvents != 0 {
		result.Triggers.MaxEvents = override.Triggers.MaxEvents
	}
	if len(override.Triggers.NarrativeTriggers) > 0 {
		result.Triggers.NarrativeTriggers = override.Triggers.NarrativeTriggers
	}

	if override.Snapshot.IntervalEvents != 0 {
		result.Snapshot.IntervalEvents = override.Snapshot.IntervalEvents
	}
	if override.Snapshot.IntervalMs != 0 {
		result.Snapshot.IntervalMs = override.Snapshot.IntervalMs
	}
	if override.Snapshot.MinioPath != "" {
		result.Snapshot.MinioPath = override.Snapshot.MinioPath
	}

	return &result
}

// MergeProfiles объединяет базовый профиль и переопределение.
// Экспортируемая версия для orchestrator.go.
func MergeProfiles(base, override *Profile) *Profile {
	return mergeProfiles(base, override)
}
