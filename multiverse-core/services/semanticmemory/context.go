// Package semanticmemory handles event processing and indexing.
package semanticmemory

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// BuildEventBasedContext creates a context string based on recent events
func (i *Indexer) BuildEventBasedContext(ctx context.Context, eventType string, limit int) (string, error) {
	// Retrieve events from storage
	events, err := i.GetEventsByType(ctx, eventType, limit)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve events: %w", err)
	}

	// Build context from events
	var contextParts []string
	for _, eventText := range events {
		contextParts = append(contextParts, eventText)
	}

	return strings.Join(contextParts, "\n\n"), nil
}

// GetContextWithEvents retrieves full context for entity IDs and includes relevant events
func (i *Indexer) GetContextWithEvents(ctx context.Context, entityIDs []string, eventTypes []string, depth int) (map[string]string, error) {
	// Get entity context
	entityContexts, err := i.GetContext(ctx, entityIDs, depth)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve entity context: %w", err)
	}

	// Enhance context with events
	for _, eventType := range eventTypes {
		eventContext, err := i.BuildEventBasedContext(ctx, eventType, 5) // Limit to 5 recent events
		if err != nil {
			log.Printf("Warning: failed to build event context for type %s: %v", eventType, err)
			continue
		}

		// Add event context to each entity context
		for entityID, entityContext := range entityContexts {
			enhancedContext := fmt.Sprintf("%s\n\nRecent %s events:\n%s", entityContext, eventType, eventContext)
			entityContexts[entityID] = enhancedContext
		}
	}

	return entityContexts, nil
}
