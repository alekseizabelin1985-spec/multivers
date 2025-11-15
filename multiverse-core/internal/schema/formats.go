// Package schema defines custom JSON Schema formats for the Multiverse.
package schema

import (
	"regexp"

	"github.com/google/uuid"
	"github.com/xeipuuv/gojsonschema"
)

// eventIDFormatChecker implements gojsonschema.FormatChecker for event_id.
type eventIDFormatChecker struct{}

// IsFormat validates that the input is a valid UUID.
func (c eventIDFormatChecker) IsFormat(input interface{}) bool {
	if s, ok := input.(string); ok {
		_, err := uuid.Parse(s)
		return err == nil
	}
	return false
}

// entityIDFormatChecker implements gojsonschema.FormatChecker for entity_id.
type entityIDFormatChecker struct{}

// IsFormat validates that the input is a valid entity ID (UUID or semantic).
func (c entityIDFormatChecker) IsFormat(input interface{}) bool {
	if s, ok := input.(string); ok {
		if len(s) == 0 {
			return false
		}
		// Accept UUIDs
		if _, err := uuid.Parse(s); err == nil {
			return true
		}
		// Accept semantic IDs: letters, digits, hyphens, underscores, dots
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9._-]+$`, s)
		return matched
	}
	return false
}

// RegisterCustomFormats registers event_id and entity_id formats.
func RegisterCustomFormats() {
	gojsonschema.FormatCheckers.Add("event_id", eventIDFormatChecker{})
	gojsonschema.FormatCheckers.Add("entity_id", entityIDFormatChecker{})
}
