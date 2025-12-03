// Package semanticmemory implements the SemanticMemory service.
package semanticmemory

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"multiverse-core/internal/eventbus"

	"github.com/gorilla/mux"
)

// Service manages the SemanticMemory lifecycle.
type Service struct {
	bus     *eventbus.EventBus
	indexer *Indexer
	server  *http.Server
}

// contextRequest represents a context request.
type contextRequest struct {
	EntityIDs []string `json:"entity_ids"`
	Depth     int      `json:"depth,omitempty"`
}

// eventRequest represents an event request.
type eventRequest struct {
	EventType string `json:"event_type"`
	Limit     int    `json:"limit,omitempty"`
}

// contextWithEventsRequest represents a context request with events.
type contextWithEventsRequest struct {
	EntityIDs  []string `json:"entity_ids"`
	EventTypes []string `json:"event_types"`
	Depth      int      `json:"depth,omitempty"`
}

// NewService creates a new SemanticMemory service.
func NewService(bus *eventbus.EventBus) (*Service, error) {
	indexer, err := NewIndexer()
	if err != nil {
		return nil, err
	}

	// Setup HTTP server
	r := mux.NewRouter()
	r.HandleFunc("/v1/context", func(w http.ResponseWriter, r *http.Request) {
		var req contextRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		contexts, err := indexer.GetContext(r.Context(), req.EntityIDs, req.Depth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{"contexts": contexts}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(response)
	}).Methods("POST")

	// Add new endpoint for events
	r.HandleFunc("/v1/events", func(w http.ResponseWriter, r *http.Request) {
		var req eventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Set default limit if not provided
		if req.Limit == 0 {
			req.Limit = 10
		}

		// Retrieve events from storage
		events, err := indexer.GetEventsByType(r.Context(), req.EventType, req.Limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{"events": events}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(response)
	}).Methods("POST")

	// Add new endpoint for context with events
	r.HandleFunc("/v1/context-with-events", func(w http.ResponseWriter, r *http.Request) {
		var req contextWithEventsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log.Println(req)
		// Retrieve context with events from storage
		contexts, err := indexer.GetContextWithEvents(r.Context(), req.EntityIDs, req.EventTypes, req.Depth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{"contexts": contexts}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(response)
	}).Methods("POST")

	semanticport := os.Getenv("SEMANTIC_PORT")
	if semanticport == "" {
		semanticport = "8080"
	}

	server := &http.Server{
		Addr:         ":" + semanticport,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &Service{
		bus:     bus,
		indexer: indexer,
		server:  server,
	}, nil
}

// Run starts the service and blocks until context is cancelled.
func (s *Service) Run(ctx context.Context) error {
	// Start HTTP server
	go func() {
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server failed: %v", err)
		}
	}()

	// Subscribe to all event topics for comprehensive context
	go s.bus.Subscribe(ctx, eventbus.TopicPlayerEvents, "semantic-memory-player-group", s.indexer.HandleEvent)
	go s.bus.Subscribe(ctx, eventbus.TopicWorldEvents, "semantic-memory-world-group", s.indexer.HandleEvent)
	go s.bus.Subscribe(ctx, eventbus.TopicGameEvents, "semantic-memory-game-group", s.indexer.HandleEvent)
	go s.bus.Subscribe(ctx, eventbus.TopicSystemEvents, "semantic-memory-system-group", s.indexer.HandleEvent)
	go s.bus.Subscribe(ctx, eventbus.TopicScopeManagement, "semantic-memory-scope-group", s.indexer.HandleEvent)
	go s.bus.Subscribe(ctx, eventbus.TopicNarrativeOutput, "semantic-memory-narrative-group", s.indexer.HandleEvent)

	<-ctx.Done()

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	s.server.Shutdown(shutdownCtx)

	// Close chroma client
	//s.indexer.Close()

	// Close Neo4j driver
	s.indexer.neo4j.Close()

	return ctx.Err()
}
