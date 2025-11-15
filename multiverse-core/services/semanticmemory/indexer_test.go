// Package semanticmemory handles event processing and indexing.
package semanticmemory

import (
	"context"
	"testing"
	"time"

	"multiverse-core/internal/eventbus"
)

func TestIndexer_HandleEvent(t *testing.T) {
	// Create a new indexer
	indexer, err := NewIndexer()
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}
	defer indexer.neo4j.Close()

	// Create a test event
	event := eventbus.Event{
		EventID:   "test-event-1",
		EventType: "player.action.move",
		Timestamp: time.Now(),
		Source:    "player-1",
		WorldID:   "test-world",
		ScopeID:   stringPtr("test-scope"),
		Payload: map[string]interface{}{
			"action": "move",
			"from":   "location-a",
			"to":     "location-b",
		},
	}

	// Handle the event
	indexer.HandleEvent(event)

	// Verify that the event was saved
	ctx := context.Background()
	events, err := indexer.GetEventsByType(ctx, "player.action.move", 10)
	if err != nil {
		t.Fatalf("Failed to retrieve events: %v", err)
	}

	if len(events) == 0 {
		t.Error("No events were retrieved")
	}
}

func TestIndexer_GetContextWithEvents(t *testing.T) {
	// Create a new indexer
	indexer, err := NewIndexer()
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}
	defer indexer.neo4j.Close()

	// Create a test event
	event := eventbus.Event{
		EventID:   "test-event-2",
		EventType: "player.action.attack",
		Timestamp: time.Now(),
		Source:    "player-1",
		WorldID:   "test-world",
		ScopeID:   stringPtr("test-scope"),
		Payload: map[string]interface{}{
			"action": "attack",
			"target": "enemy-1",
			"damage": 10,
		},
	}

	// Handle the event
	indexer.HandleEvent(event)

	// Get context with events
	ctx := context.Background()
	contexts, err := indexer.GetContextWithEvents(ctx, []string{"player-1"}, []string{"player.action.attack"}, 1)
	if err != nil {
		t.Fatalf("Failed to get context with events: %v", err)
	}

	if len(contexts) == 0 {
		t.Error("No contexts were retrieved")
	}
}

func stringPtr(s string) *string {
	return &s
}