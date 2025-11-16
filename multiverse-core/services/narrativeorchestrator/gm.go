// services/narrativeorchestrator/gm.go

package narrativeorchestrator

import (
	"time"

	"multiverse-core/internal/spatial"
)

// HistoryEntry сохраняет событие с временной меткой.
type HistoryEntry struct {
	EventID   string    `json:"event_id"`
	Timestamp time.Time `json:"timestamp"`
}

// GMInstance represents a stateful Game Master for a scope.
type GMInstance struct {
	ScopeID         string                  `json:"scope_id"`
	ScopeType       string                  `json:"scope_type"`
	WorldID         string                  `json:"world_id"`
	VisibilityScope spatial.VisibilityScope `json:"visibility_scope"`
	State           map[string]interface{}  `json:"state"`
	History         []HistoryEntry          `json:"history"`
	Config          map[string]interface{}  `json:"config"`
	CreatedAt       time.Time               `json:"created_at"`
}
