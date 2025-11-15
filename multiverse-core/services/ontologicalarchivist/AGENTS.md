# AGENTS.md for OntologicalArchivist

This file provides guidance to AI assistants when working with the OntologicalArchivist service.

## Service Overview

OntologicalArchivist handles storage and management of ontological schemas and entities.

### Key Responsibilities
- Storing versioned entity schemas
- Supporting ontological integrity
- Providing schema access to other services
- Managing schema versions

### Event Integration
- Subscribes to: `system_events` topic with types `schema.save`, `schema.get`
- Publishes: schema storage results
- Event types: schema save/get events

## Build/Run Commands

- Build: `make build-service SERVICE=ontological-archivist`
- Run: `make run SERVICE=ontological-archivist`
- Logs: `make logs-service SERVICE=ontological-archivist`

## Code Style Guidelines

- All services written in Go 1.25
- Use JSON Schema Draft 7 for payload validation
- Follow event-driven architecture with Kafka/Redpanda
- Use entity paths with dot notation for nested access (e.g., `payload.health.current`)
- Build with CGO disabled (`CGO_ENABLED=0`)
- Use UUIDs for entity IDs
- Containerize with Docker multi-stage builds

## Key Patterns and Utilities

- Event bus integration via `internal/eventbus`
- Versioned schema storage in MinIO
- Schema validation support
- Stateless service architecture

## Directory Structure

- Service implementation: `services/ontologicalarchivist/`
- Command entry point: `cmd/ontological-archivist/
- Internal packages: `internal/eventbus/`, `internal/schema/`

## Service-Specific Conventions

- Stores schemas in MinIO with versioning
- Uses path pattern: `schemas/{type}/{name}/v{version}.json`
- Maintains schema version history
- Integrates with WorldGenerator for schema retrieval
- Works with UniverseGenesisOracle for schema storage
- Supports EntityManager for entity schemas
- Communicates with SemanticMemory for indexing context
- Backward compatibility support
- HTTP API on port 8080 (default)