// Package worldgenerator handles communication with OntologicalArchivist.
package worldgenerator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// ArchivistClient communicates with OntologicalArchivist.
type ArchivistClient struct {
	BaseURL string
}

// SchemaRequest represents a schema save request.
type SchemaRequest struct {
	SchemaType string          `json:"schema_type"`
	Name       string          `json:"name"`
	Version    string          `json:"version"`
	Schema     json.RawMessage `json:"schema"`
}

// NewArchivistClient creates a new ArchivistClient.
func NewArchivistClient() *ArchivistClient {
	url := os.Getenv("ARCHIVIST_URL")
	if url == "" {
		url = "http://ontological-archivist:8081"
	}
	return &ArchivistClient{BaseURL: url}
}

// SaveSchema saves a schema to OntologicalArchivist.
func (ac *ArchivistClient) SaveSchema(ctx context.Context, schemaType, name, version string, schemaData []byte) error {
	reqBody, _ := json.Marshal(SchemaRequest{
		SchemaType: schemaType,
		Name:       name,
		Version:    version,
		Schema:     schemaData,
	})

	log.Println(ac.BaseURL + "/v1/schemas")
	log.Println(string(reqBody))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(ac.BaseURL+"/v1/schemas", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("archivist connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("archivist returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
