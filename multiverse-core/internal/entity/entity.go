package entity

import (
	"reflect"
	"strings"
	"time"
)

type Entity struct {
	EntityID   string                 `json:"entity_id"`
	EntityType string                 `json:"entity_type"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Payload    map[string]interface{} `json:"payload"`
	History    []HistoryEntry         `json:"history"`
}

type HistoryEntry struct {
	EventID   string    `json:"event_id"`
	Timestamp time.Time `json:"timestamp"`
}

func NewEntity(entityID, entityType string, payload map[string]interface{}) *Entity {
	if payload == nil {
		payload = make(map[string]interface{})
	}
	now := time.Now().UTC()
	return &Entity{
		EntityID:   entityID,
		EntityType: entityType,
		CreatedAt:  now,
		UpdatedAt:  now,
		Payload:    payload,
		History:    make([]HistoryEntry, 0),
	}
}

func (e *Entity) AddHistoryEntry(eventID string, timestamp time.Time) {
	e.History = append(e.History, HistoryEntry{
		EventID:   eventID,
		Timestamp: timestamp,
	})
	e.UpdatedAt = time.Now().UTC()
}

func (e *Entity) Set(key string, value interface{}) bool {
	if strings.Contains(key, ".") {
		return e.SetPath(key, value)
	}
	if current, exists := e.Payload[key]; exists {
		if reflect.DeepEqual(current, value) {
			return false
		}
	}
	e.Payload[key] = value
	e.UpdatedAt = time.Now().UTC()
	return true
}

func (e *Entity) Get(key string) (interface{}, bool) {
	if strings.Contains(key, ".") {
		return e.GetPath(key)
	}
	val, exists := e.Payload[key]
	return val, exists
}

func (e *Entity) Has(key string) bool {
	if strings.Contains(key, ".") {
		return e.HasPath(key)
	}
	_, exists := e.Payload[key]
	return exists
}

func (e *Entity) Remove(key string) bool {
	if strings.Contains(key, ".") {
		return e.RemovePath(key)
	}
	if _, exists := e.Payload[key]; exists {
		delete(e.Payload, key)
		e.UpdatedAt = time.Now().UTC()
		return true
	}
	return false
}

func (e *Entity) GetPath(path string) (interface{}, bool) {
	keys := strings.Split(path, ".")
	if len(keys) == 0 {
		return nil, false
	}
	current := interface{}(e.Payload)
	for _, key := range keys {
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[key]; exists {
				current = val
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}
	return current, true
}

func (e *Entity) HasPath(path string) bool {
	_, exists := e.GetPath(path)
	return exists
}

func (e *Entity) RemovePath(path string) bool {
	keys := strings.Split(path, ".")
	if len(keys) == 0 {
		return false
	}
	var parentMap map[string]interface{}
	var targetKey string
	current := interface{}(e.Payload)
	for i, key := range keys {
		if i == len(keys)-1 {
			targetKey = key
			if m, ok := current.(map[string]interface{}); ok {
				parentMap = m
			} else {
				return false
			}
			break
		}
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[key]; exists {
				current = val
			} else {
				return false
			}
		} else {
			return false
		}
	}
	if parentMap == nil {
		return false
	}
	if _, exists := parentMap[targetKey]; exists {
		delete(parentMap, targetKey)
		e.UpdatedAt = time.Now().UTC()
		return true
	}
	return false
}

func (e *Entity) SetPath(path string, value interface{}) bool {
	keys := strings.Split(path, ".")
	if len(keys) == 0 {
		return false
	}
	current := interface{}(e.Payload)
	for i := 0; i < len(keys)-1; i++ {
		key := keys[i]
		var next interface{}
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[key]; exists {
				next = val
			} else {
				next = make(map[string]interface{})
				m[key] = next
			}
		} else {
			return false
		}
		if _, ok := next.(map[string]interface{}); !ok {
			newMap := make(map[string]interface{})
			if m, isMap := current.(map[string]interface{}); isMap {
				m[key] = newMap
			}
			next = newMap
		}
		current = next
	}
	finalKey := keys[len(keys)-1]
	if m, ok := current.(map[string]interface{}); ok {
		if currentVal, exists := m[finalKey]; exists {
			if reflect.DeepEqual(currentVal, value) {
				return false
			}
		}
		m[finalKey] = value
		e.UpdatedAt = time.Now().UTC()
		return true
	}
	return false
}

func (e *Entity) AddToStringSlice(key, item string) bool {
	current, exists := e.Payload[key]
	if !exists {
		e.Payload[key] = []string{item}
		e.UpdatedAt = time.Now().UTC()
		return true
	}
	switch v := current.(type) {
	case []string:
		for _, s := range v {
			if s == item {
				return false
			}
		}
		e.Payload[key] = append(v, item)
	case []interface{}:
		for _, s := range v {
			if s == item {
				return false
			}
		}
		e.Payload[key] = append(v, item)
	default:
		e.Payload[key] = []string{item}
	}
	e.UpdatedAt = time.Now().UTC()
	return true
}

func (e *Entity) RemoveFromStringSlice(key, item string) bool {
	current, exists := e.Payload[key]
	if !exists {
		return false
	}
	switch v := current.(type) {
	case []string:
		for i, s := range v {
			if s == item {
				e.Payload[key] = append(v[:i], v[i+1:]...)
				e.UpdatedAt = time.Now().UTC()
				return true
			}
		}
	case []interface{}:
		for i, s := range v {
			if s == item {
				e.Payload[key] = append(v[:i], v[i+1:]...)
				e.UpdatedAt = time.Now().UTC()
				return true
			}
		}
	}
	return false
}
