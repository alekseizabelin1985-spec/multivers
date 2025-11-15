// Package entity provides helper types and constructors.

package entity

import "time"

// SetPayload updates the dynamic state of the entity.
func (e *Entity) SetPayload(payload map[string]interface{}) {
	e.Payload = payload
	e.UpdatedAt = time.Now().UTC()
}
