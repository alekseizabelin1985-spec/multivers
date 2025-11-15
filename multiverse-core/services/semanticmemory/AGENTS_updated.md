# AGENTS.md for SemanticMemory

This file provides guidance to AI assistants when working with the SemanticMemory service.

## Service Overview

SemanticMemory is a powerful memory system with dual indexing (vector + graph) for RAG.

### Key Responsibilities
- Storing and searching semantic entities
- Supporting RAG for narratives and other AI services
- Combining vector and graph indexing
- Searching by entity context and relationships
- Storing events for context and replay

### Event Integration
- Subscribes to: all event topics (`player_events`, `world_events`, `game_events`, `system_events`, `scope_management`, `narrative_output`)
- Publishes: search results and context data
- Event types: all system events (entity creation/update, world generation, player actions, etc.)

## Build/Run Commands

- Build: `make build-service SERVICE=semantic-memory`
- Run: `make run SERVICE=semantic-memory`
- Logs: `make logs-service SERVICE=semantic-memory`

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
- Dual indexing system (ChromaDB + Neo4j)
- RAG (Retrieval-Augmented Generation) support
- Stateless service architecture
- Event storage with persistence

## Directory Structure

- Service implementation: `services/semanticmemory/`
- Command entry point: `cmd/semantic-memory/`
- Internal packages: `internal/eventbus/`

## Service-Specific Conventions

- Vector indexing in ChromaDB
- Graph model in Neo4j
- Search cache management
- Integrates with NarrativeOrchestrator for context
- Works with AscensionOracle for AI context
- Supports WorldGenerator for world context
- Stateless with persistent storage
- Entity-based indexing approach
- Event-based context storage
- HTTP API on port 8082 (default)