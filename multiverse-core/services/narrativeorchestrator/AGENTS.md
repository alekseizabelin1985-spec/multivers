# AGENTS.md for NarrativeOrchestrator

This file provides guidance to AI assistants when working with the NarrativeOrchestrator (GM) service.

## Service Overview

NarrativeOrchestrator (GM) is a dynamic, stateful agent that creates immersive narratives for game areas.

### Key Responsibilities
- Generating immersive narratives for game areas
- Maintaining story state between events
- Using RAG from Semantic Memory and Ascension Oracle
- Managing lifecycle through system events

### Event Integration
- Subscribes to: `system_events`, `world_events`, `game_events` topics
- Publishes to: `narrative_output` topic
- Event types: `gm.created`, `gm.deleted`, `gm.merged`, `gm.split`

## Build/Run Commands

- Build: `make build-service SERVICE=narrative-orchestrator`
- Run: `make run SERVICE=narrative-orchestrator`
- Logs: `make logs-service SERVICE=narrative-orchestrator`

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
- Dynamic agent creation/deletion based on events
- RAG (Retrieval-Augmented Generation) integration with Semantic Memory
- Stateful narrative context management

## Directory Structure

- Service implementation: `services/narrativeorchestrator/`
- Command entry point: `cmd/narrative-orchestrator/`
- Internal packages: `internal/eventbus/`

## Service-Specific Conventions

- GM instances stored in memory: `map[scope_id]*GMInstance`
- Uses scope-based narrative generation
- Integrates with Semantic Memory Builder for entity context
- Communicates with Ascension Oracle for narrative generation
- Supports multi-step story arcs
- Event group subscriptions: `narrative-orchestrator-group`, `narrative-world-group`, `narrative-game-group`