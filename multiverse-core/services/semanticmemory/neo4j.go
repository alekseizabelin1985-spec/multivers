// Package semanticmemory handles Neo4j integration.
package semanticmemory

import (
	"context"
	"fmt"
	"os"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jClient handles communication with Neo4j.
type Neo4jClient struct {
	driver neo4j.Driver
}

// NewNeo4jClient creates a new Neo4jClient.
func NewNeo4jClient() (*Neo4jClient, error) {
	uri := os.Getenv("NEO4J_URI")
	if uri == "" {
		uri = "neo4j://neo4j:7687"
	}
	user := os.Getenv("NEO4J_USER")
	if user == "" {
		user = "neo4j"
	}
	password := os.Getenv("NEO4J_PASSWORD")
	if password == "" {
		password = "password"
	}

	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, fmt.Errorf("neo4j driver creation failed: %w", err)
	}

	// Test connection (no context in v5)
	if err := driver.VerifyConnectivity(); err != nil {
		driver.Close()
		return nil, fmt.Errorf("neo4j connectivity test failed: %w", err)
	}

	return &Neo4jClient{driver: driver}, nil
}

// UpsertEntity creates or updates an entity node in Neo4j.
func (n *Neo4jClient) UpsertEntity(ctx context.Context, entityID, entityType string, payload map[string]interface{}) error {
	// В v5: NewSession принимает только SessionConfig (без context)
	session := n.driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close()

	query := `
	MERGE (e:Entity {id: $entity_id})
	SET e.type = $entity_type,
	    e += $payload
	RETURN e
	`

	// В v5: Run принимает context как первый аргумент
	_, err := session.Run(query, map[string]interface{}{
		"entity_id":   entityID,
		"entity_type": entityType,
		"payload":     payload,
	})
	return err
}

// CreateRelationship creates a relationship between entities.
func (n *Neo4jClient) CreateRelationship(ctx context.Context, fromID, toID, relType string) error {
	session := n.driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close()

	query := fmt.Sprintf(`
	MATCH (a:Entity {id: $from_id})
	MATCH (b:Entity {id: $to_id})
	MERGE (a)-[r:%s]->(b)
	RETURN r
	`, relType)

	_, err := session.Run(query, map[string]interface{}{
		"from_id": fromID,
		"to_id":   toID,
	})
	return err
}

// StoreEvent stores an event in Neo4j as a node with relationships to relevant entities.
func (n *Neo4jClient) StoreEvent(ctx context.Context, eventID string, eventType string, timestamp string, entityID string, eventData map[string]interface{}) error {
	session := n.driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		// Create or merge the event node with all event data
		eventQuery := `
			MERGE (e:Event {id: $eventID})
			SET e.type = $eventType, e.timestamp = $timestamp
			SET e += $eventData
			RETURN e
		`

		_, err := tx.Run(eventQuery, map[string]interface{}{
			"eventID":   eventID,
			"eventType": eventType,
			"timestamp": timestamp,
			"eventData": eventData,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create/merge event node: %w", err)
		}

		// Create relationship to the associated entity if entityID is provided
		if entityID != "" {
			relQuery := `
				MERGE (ev:Event {id: $eventID})
				MERGE (en:Entity {id: $entityID})
				MERGE (ev)-[:RELATED_TO]->(en)
			`

			_, err := tx.Run(relQuery, map[string]interface{}{
				"eventID":  eventID,
				"entityID": entityID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create relationship to entity: %w", err)
			}
		}

		return nil, nil
	})

	return err
}

// GetEventsByType retrieves events of a specific type from Neo4j, sorted by timestamp.
func (n *Neo4jClient) GetEventsByType(ctx context.Context, eventType string, limit int) ([]map[string]interface{}, error) {
	session := n.driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		query := `
			MATCH (e:Event {type: $eventType})
			RETURN e
			ORDER BY e.timestamp DESC
			LIMIT $limit
		`

		records, err := tx.Run(query, map[string]interface{}{
			"eventType": eventType,
			"limit":     limit,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}

		var events []map[string]interface{}
		for records.Next() {
			record := records.Record()
			eventNode, ok := record.Get("e")
			if !ok {
				continue // Skip invalid records
			}
			
			// Convert node to map
			if node, ok := eventNode.(neo4j.Node); ok {
				events = append(events, node.Props)
			}
		}

		return events, nil
	})

	if err != nil {
		return nil, err
	}

	events, ok := result.([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert result to events slice")
	}

	return events, nil
}

// Close closes the Neo4j driver.
func (n *Neo4jClient) Close() {
	n.driver.Close()
}
