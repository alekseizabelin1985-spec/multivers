package narrativeorchestrator

import (
	"time"
)

// HistoryEntry сохраняет событие с временной меткой.
type HistoryEntry struct {
	EventID   string    `json:"event_id"`
	Timestamp time.Time `json:"timestamp"`
}

// GMInstance represents a stateful Game Master for a scope.
type GMInstance struct {
	ScopeID   string                 `json:"scope_id"`
	ScopeType string                 `json:"scope_type"`
	WorldID   string                 `json:"world_id"`
	State     map[string]interface{} `json:"state"` // Для KnowledgeBase: entities, canon, last_mood
	History   []HistoryEntry         `json:"history"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt time.Time              `json:"created_at"`
}