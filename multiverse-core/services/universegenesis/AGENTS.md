# AGENTS.md for UniverseGenesisOracle

This file provides guidance to AI assistants when working with the UniverseGenesisOracle service.

## Service Overview

UniverseGenesisOracle is a specialized AI service for generating universes with philosophical integrity.

### Key Responsibilities
- Generating new universes with philosophical integrity
- Creating world structure and parameters
- Supporting semantic depth in generation
- Generating schemas and structures for WorldGenerator

### Event Integration
- Subscribes to: `system_events` topic with type `universe.generate`
- Publishes to: `world_events` topic
- Event types: universe generation requests

## Build/Run Commands

- Build: `make build-service SERVICE=universe-genesis-oracle`
- Run: `make run SERVICE=universe-genesis-oracle`
- Logs: `make logs-service SERVICE=universe-genesis-oracle`

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
- Stateless service architecture
- Integration with AscensionOracle for generation
- AI-powered universe generation

## Directory Structure

- Service implementation: `services/universegenesis/`
- Command entry point: `cmd/universe-genesis-oracle/
- Internal packages: `internal/eventbus/`, `internal/oracle/`

## Service-Specific Conventions

- Uses Qwen3 for generation
- Ensures semantic depth and philosophical integrity
- Flexible generation of different universe types
- Integrates with WorldGenerator for structure delivery
- Works with OntologicalArchivist for schema storage
- Communicates with SemanticMemory for generation context
- Collaborates with AscensionOracle for philosophical integrity
- Generates fundamental universe hierarchy (Core, Laws)
- Creates ontological profile for Cosmic Ban
- Stateless with HTTP-based AI integration