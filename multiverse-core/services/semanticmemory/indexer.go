// Package semanticmemory handles event processing and indexing.
package semanticmemory

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"multiverse-core/internal/eventbus"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// toStringSlice converts an interface{} to a []string
func toStringSlice(v interface{}) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []interface{}:
		var result []string
		for _, item := range t {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	default:
		return []string{}
	}
}

// BuildTextContext creates a text context from entity data
func BuildTextContext(entityID, entityType string, payload map[string]interface{}) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Entity ID: %s", entityID))
	parts = append(parts, fmt.Sprintf("Entity Type: %s", entityType))

	// Add payload information
	if len(payload) > 0 {
		parts = append(parts, "Payload:")
		for key, value := range payload {
			// Convert value to string representation
			valueStr := ""
			switch v := value.(type) {
			case string:
				valueStr = v
			case []string:
				valueStr = strings.Join(v, ", ")
			case []interface{}:
				var items []string
				for _, item := range v {
					items = append(items, fmt.Sprintf("%v", item))
				}
				valueStr = strings.Join(items, ", ")
			default:
				valueStr = fmt.Sprintf("%v", v)
			}
			parts = append(parts, fmt.Sprintf("  %s: %s", key, valueStr))
		}
	}

	return strings.Join(parts, "\n")
}

// Indexer processes entity events and indexes them.
type Indexer struct {
	chroma *ChromaClient
	neo4j  *Neo4jClient
}

// NewIndexer creates a new Indexer.
func NewIndexer() (*Indexer, error) {
	chroma := NewChromaClient() // ← Использует новый клиент
	//if err != nil {
	//	return nil, err
	//}

	neo4j, err := NewNeo4jClient()
	if err != nil {
		return nil, err
	}

	return &Indexer{
		chroma: chroma,
		neo4j:  neo4j,
	}, nil
}

// HandleEvent processes all events and indexes them.
func (i *Indexer) HandleEvent(ev eventbus.Event) {
	// Process all events, not just entity-related events
	ctx := context.Background()

	// Save event to storage
	if err := i.saveEvent(ctx, ev); err != nil {
		log.Printf("Failed to save event %s: %v", ev.EventID, err)
		return
	}

	// Process entity-related events for backward compatibility
	if ev.EventType == "entity.created" || ev.EventType == "entity.updated" {
		i.processEntityEvent(ctx, ev)
	}
}

// saveEvent saves an event to both ChromaDB and Neo4j for context and replay
func (i *Indexer) saveEvent(ctx context.Context, ev eventbus.Event) error {
	// Create a text representation of the event for ChromaDB
	eventText := i.buildEventTextContext(ev)

	// Metadata for ChromaDB
	metadata := map[string]interface{}{
		"event_id":   ev.EventID,
		"event_type": ev.EventType,
		"world_id":   ev.WorldID,
		"source":     ev.Source,
		"timestamp":  ev.Timestamp,
	}

	// Add scope_id to metadata if present
	if ev.ScopeID != nil {
		metadata["scope_id"] = *ev.ScopeID
	}

	// Save to ChromaDB
	eventID := fmt.Sprintf("event_%s", ev.EventID)
	if err := i.chroma.UpsertDocument(ctx, eventID, eventText, metadata); err != nil {
		return fmt.Errorf("ChromaDB upsert failed for event %s: %w", ev.EventID, err)
	}

	// Save to Neo4j
	if err := i.saveEventToNeo4j(ctx, ev); err != nil {
		return fmt.Errorf("Neo4j upsert failed for event %s: %w", ev.EventID, err)
	}

	log.Printf("Saved event %s of type %s", ev.EventID, ev.EventType)
	return nil
}

// buildEventTextContext creates a human-readable context string from an event
func (i *Indexer) buildEventTextContext(ev eventbus.Event) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Event ID: %s", ev.EventID))
	parts = append(parts, fmt.Sprintf("Event Type: %s", ev.EventType))
	parts = append(parts, fmt.Sprintf("Timestamp: %s", ev.Timestamp.Format("2006-01-02 15:04:05")))
	parts = append(parts, fmt.Sprintf("Source: %s", ev.Source))
	parts = append(parts, fmt.Sprintf("World ID: %s", ev.WorldID))

	if ev.ScopeID != nil {
		parts = append(parts, fmt.Sprintf("Scope ID: %s", *ev.ScopeID))
	}

	// Add payload information
	if len(ev.Payload) > 0 {
		parts = append(parts, "Payload:")
		for key, value := range ev.Payload {
			// Convert value to string representation
			valueStr := ""
			switch v := value.(type) {
			case string:
				valueStr = v
			case []string:
				valueStr = strings.Join(v, ", ")
			case []interface{}:
				var items []string
				for _, item := range v {
					items = append(items, fmt.Sprintf("%v", item))
				}
				valueStr = strings.Join(items, ", ")
			default:
				valueStr = fmt.Sprintf("%v", v)
			}
			parts = append(parts, fmt.Sprintf("  %s: %s", key, valueStr))
		}
	}

	return strings.Join(parts, "\n")
}

// saveEventToNeo4j saves an event to Neo4j
func (i *Indexer) saveEventToNeo4j(ctx context.Context, ev eventbus.Event) error {
	session := i.neo4j.driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close()

	// Convert event to JSON for storage
	eventJSON, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	query := `
	MERGE (e:Event {id: $event_id})
	SET e.type = $event_type,
	    e.timestamp = $timestamp,
	    e.source = $source,
	    e.world_id = $world_id,
	    e.scope_id = $scope_id,
	    e.payload = $payload,
	    e.raw_data = $raw_data
	RETURN e
	`

	_, err = session.Run(query, map[string]interface{}{
		"event_id":   ev.EventID,
		"event_type": ev.EventType,
		"timestamp":  ev.Timestamp,
		"source":     ev.Source,
		"world_id":   ev.WorldID,
		"scope_id":   ev.ScopeID,
		"payload":    ev.Payload,
		"raw_data":   string(eventJSON),
	})
	return err
}

// processEntityEvent handles entity-related events for backward compatibility
func (i *Indexer) processEntityEvent(ctx context.Context, ev eventbus.Event) {
	entityID, ok := ev.Payload["entity_id"].(string)
	if !ok {
		log.Printf("Invalid entity event: missing entity_id")
		return
	}

	entityType, _ := ev.Payload["entity_type"].(string)
	payload, _ := ev.Payload["payload"].(map[string]interface{})

	// Build text context
	textContext := BuildTextContext(entityID, entityType, payload)

	// Metadata for ChromaDB
	metadata := map[string]interface{}{
		"entity_id":   entityID,
		"entity_type": entityType,
		"world_id":    ev.WorldID,
	}

	// Index in ChromaDB
	if err := i.chroma.UpsertDocument(ctx, entityID, textContext, metadata); err != nil {
		log.Printf("ChromaDB upsert failed for %s: %v", entityID, err)
	}

	// Index in Neo4j
	if err := i.neo4j.UpsertEntity(ctx, entityID, entityType, payload); err != nil {
		log.Printf("Neo4j upsert failed for %s: %v", entityID, err)
	}

	// Create relationships for inventory items
	if inv, ok := payload["inventory"]; ok {
		inventory := toStringSlice(inv)
		for _, itemID := range inventory {
			if err := i.neo4j.CreateRelationship(ctx, entityID, itemID, "CONTAINS"); err != nil {
				log.Printf("Neo4j relationship failed: %s -> %s: %v", entityID, itemID, err)
			}
		}
	}

	log.Printf("Indexed entity %s", entityID)
}

// GetContext retrieves full context for entity IDs.
func (i *Indexer) GetContext(ctx context.Context, entityIDs []string, depth int) (map[string]string, error) {
	// For simplicity, we use ChromaDB for text context
	return i.chroma.GetDocuments(ctx, entityIDs)
}

// GetEventsByType retrieves events by type from ChromaDB
func (i *Indexer) GetEventsByType(ctx context.Context, eventType string, limit int) ([]string, error) {
	return i.chroma.SearchEventsByType(ctx, eventType, limit)
}

// Close the chroma client when indexer is done
//func (i *Indexer) Close() error {
//if i.chroma != nil {
//	return i.chroma.Close()
//}
//return nil
//}
