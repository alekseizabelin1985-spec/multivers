package eventbus

import (
	"time"
	"github.com/google/uuid"
)

type Event struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	WorldID   string                 `json:"world_id"`
	ScopeID   *string                `json:"scope_id,omitempty"`
	Payload   map[string]interface{} `json:"payload"`
}

func NewEvent(eventType, source, worldID string, payload map[string]interface{}) Event {
	if payload == nil {
		payload = make(map[string]interface{})
	}
	return Event{
		EventID:   uuid.NewString(),
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Source:    source,
		WorldID:   worldID,
		Payload:   payload,
	}
}
