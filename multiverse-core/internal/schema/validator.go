// Package schema provides JSON Schema validation with custom formats.
package schema

import (
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// Validator validates data against a JSON Schema.
type Validator struct {
	schemaLoader gojsonschema.JSONLoader
}

// NewValidator creates a validator from schema bytes.
// Make sure to call RegisterCustomFormats() before this.
func NewValidator(schemaData []byte) (*Validator, error) {
	schemaLoader := gojsonschema.NewBytesLoader(schemaData)
	return &Validator{schemaLoader: schemaLoader}, nil
}

// Validate validates a map[string]interface{} against the schema.
func (v *Validator) Validate(data map[string]interface{}) error {
	documentLoader := gojsonschema.NewGoLoader(data)
	result, err := gojsonschema.Validate(v.schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		return fmt.Errorf("validation failed: %v", errors)
	}
	return nil
}

// ValidateBytes validates raw JSON bytes.
func (v *Validator) ValidateBytes(data []byte) error {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return v.Validate(obj)
}
