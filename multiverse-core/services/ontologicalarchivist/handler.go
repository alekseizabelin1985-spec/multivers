// Package ontologicalarchivist handles HTTP requests.
package ontologicalarchivist

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// schemaRequest represents a schema save request.
type schemaRequest struct {
	SchemaType string          `json:"schema_type"`
	Name       string          `json:"name"`
	Version    string          `json:"version"`
	Schema     json.RawMessage `json:"schema"`
}

// NewBytesReader creates a new io.Reader from bytes.
func NewBytesReader(data []byte) io.Reader {
	return bytes.NewReader(data)
}

// ReadAll reads all data from an io.Reader.
func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

// handleSaveSchema handles POST /v1/schemas
func (s *Service) handleSaveSchema(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var req schemaRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.SchemaType == "" || req.Name == "" || req.Version == "" {
		http.Error(w, "Missing required fields: schema_type, name, version", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.SaveSchema(ctx, req.SchemaType, req.Name, req.Version, req.Schema); err != nil {
		log.Printf("Save schema failed: %v", err)
		http.Error(w, "Failed to save schema", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status": "success"}`))
}

// handleGetSchema handles GET /v1/schemas/{schema_type}/{name}/{version}
func (s *Service) handleGetSchema(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	schemaType := vars["schema_type"]
	name := vars["name"]
	version := vars["version"]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	schemaData, err := s.GetSchema(ctx, schemaType, name, version)
	if err != nil {
		log.Printf("Get schema failed: %v", err)
		http.Error(w, "Schema not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(schemaData)
}

// SetupRoutes sets up HTTP routes.
func (s *Service) SetupRoutes(r *mux.Router) {
	r.HandleFunc("/v1/schemas", s.handleSaveSchema).Methods("POST")
	r.HandleFunc("/v1/schemas/{schema_type}/{name}/{version}", s.handleGetSchema).Methods("GET")
}
