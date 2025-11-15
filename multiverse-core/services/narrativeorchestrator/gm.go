// services/narrativeorchestrator/gm.go
package narrativeorchestrator

import "time"

// GMInstance represents a stateful Game Master for a scope.
type GMInstance struct {
	ScopeID   string                 `json:"scope_id"`
	ScopeType string                 `json:"scope_type"`
	WorldID   string                 `json:"world_id"`
	State     map[string]interface{} `json:"state"`
	History   []string               `json:"history"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt time.Time              `json:"created_at"`
}
