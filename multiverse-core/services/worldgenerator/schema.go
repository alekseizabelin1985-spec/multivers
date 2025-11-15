// Package worldgenerator handles schema generation.
package worldgenerator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// BaseEntitySchema is the base schema for all entities.
const BaseEntitySchema = `{
	"type": "object",
	"properties": {
		"entity_id": {"type": "string", "format": "entity_id"},
		"entity_type": {"type": "string"},
		"created_at": {"type": "string", "format": "date-time"},
		"updated_at": {"type": "string", "format": "date-time"},
		"payload": {"type": "object"},
		"history": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"event_id": {"type": "string", "format": "event_id"},
					"timestamp": {"type": "string", "format": "date-time"}
				},
				"required": ["event_id", "timestamp"],
				"additionalProperties": false
			}
		}
	},
	"required": ["entity_id", "entity_type", "created_at", "updated_at", "payload", "history"]
}`

// GenerateEntitySchema generates and saves a schema for an entity type.
func (wg *WorldGenerator) GenerateEntitySchema(ctx context.Context, entityType, worldSeed string) error {
	log.Printf("Generating schema for entity type: %s", entityType)

	// Generate payload schema via Oracle
	payloadSchemaStr, err := wg.generatePayloadSchema(ctx, entityType, worldSeed)
	if err != nil {
		return fmt.Errorf("payload schema generation failed: %w", err)
	}

	// Parse base schema
	var baseSchema map[string]interface{}
	if err := json.Unmarshal([]byte(BaseEntitySchema), &baseSchema); err != nil {
		return fmt.Errorf("base schema parse failed: %w", err)
	}

	// Parse payload schema
	var payloadSchema map[string]interface{}
	if err := json.Unmarshal([]byte(payloadSchemaStr), &payloadSchema); err != nil {
		return fmt.Errorf("payload schema parse failed: %w", err)
	}

	// Merge schemas
	properties := baseSchema["properties"].(map[string]interface{})
	properties["payload"] = payloadSchema

	// Convert to bytes
	fullSchemaBytes, err := json.Marshal(baseSchema)
	if err != nil {
		return fmt.Errorf("schema marshal failed: %w", err)
	}

	// Save to OntologicalArchivist
	if err := wg.archivist.SaveSchema(ctx, "entity", entityType, "1.0", fullSchemaBytes); err != nil {
		return fmt.Errorf("failed to save schema to archivist: %w", err)
	}

	log.Printf("Schema for %s saved to Archivist", entityType)
	return nil
}

// GenerateEntitySchemaWithArchivist generates and saves a schema for an entity type using provided archivist client
func GenerateEntitySchemaWithArchivist(archivist *ArchivistClient, ctx context.Context, entityType, worldSeed string) error {
	log.Printf("Generating schema for entity type: %s", entityType)

	// Create temporary WorldGenerator to use helper methods
	// We need to create a copy of the archivist to avoid modifying the original
	wg := &WorldGenerator{archivist: *archivist}
	
	// Generate payload schema via Oracle
	payloadSchemaStr, err := wg.generatePayloadSchema(ctx, entityType, worldSeed)
	if err != nil {
		return fmt.Errorf("payload schema generation failed: %w", err)
	}

	// Parse base schema
	var baseSchema map[string]interface{}
	if err := json.Unmarshal([]byte(BaseEntitySchema), &baseSchema); err != nil {
		return fmt.Errorf("base schema parse failed: %w", err)
	}

	// Parse payload schema
	var payloadSchema map[string]interface{}
	if err := json.Unmarshal([]byte(payloadSchemaStr), &payloadSchema); err != nil {
		return fmt.Errorf("payload schema parse failed: %w", err)
	}

	// Merge schemas
	properties := baseSchema["properties"].(map[string]interface{})
	properties["payload"] = payloadSchema

	// Convert to bytes
	fullSchemaBytes, err := json.Marshal(baseSchema)
	if err != nil {
		return fmt.Errorf("schema marshal failed: %w", err)
	}

	// Save to OntologicalArchivist
	if err := archivist.SaveSchema(ctx, "entity", entityType, "1.0", fullSchemaBytes); err != nil {
		return fmt.Errorf("failed to save schema to archivist: %w", err)
	}

	log.Printf("Schema for %s saved to Archivist", entityType)
	return nil
}

// generatePayloadSchema asks Ascension Oracle to generate a payload schema.
func (wg *WorldGenerator) generatePayloadSchema(ctx context.Context, entityType, worldSeed string) (string, error) {
	prompt := fmt.Sprintf(`
 Сгенерируй ТОЛЬКО JSON Schema Draft 7 для поля "payload" сущности типа "%s" в мире с семенем "%s".
 
 Требования:
 1. Строго в формате JSON Schema.
 2. Используй "format": "entity_id" для ссылок на другие сущности.
 3. Укажи "required" поля.
 4. Пример для игрока:
 {
   "type": "object",
   "properties": {
     "name": {"type": "string"},
     "hp": {"type": "integer", "minimum": 0},
     "inventory": {
       "type": "array",
       "items": {"type": "string", "format": "entity_id"}
     }
   },
   "required": ["name", "hp"]
 }
 
 Верни ТОЛЬКО JSON без пояснений.
 `, entityType, worldSeed)

	resp, err := CallOracle(ctx, prompt)
	if err != nil {
		return "", err
	}

	return resp, nil
}
