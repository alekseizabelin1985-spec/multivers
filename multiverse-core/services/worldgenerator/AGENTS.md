# AGENTS.md for WorldGenerator

This file provides guidance to AI assistants when working with the WorldGenerator service.

## Service Overview

WorldGenerator creates and manages game worlds using Ascension Oracle for generative content.

### Key Responsibilities
- Generating game worlds and their structure
- Creating schemas and entities for worlds
- Supporting different world types and parameters
- Integrating with OntologicalArchivist for schema storage
- Generating regions with different biomes (forests, mountains, plains, etc.)
- Generating water bodies (rivers, seas, lakes)
- Generating cities with basic characteristics

### Event Integration
- Subscribes to: `world_events` topic with type `world.generate`
- Publishes to: `world_events` and `system_events` topics
- Event types: `world.generated`, `world.request.generated`, `entity.created` (for regions/cities/water)

## Build/Run Commands

- Build: `make build-service SERVICE=world-generator`
- Run: `make run SERVICE=world-generator`
- Logs: `make logs-service SERVICE=world-generator`

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
- Schema management with OntologicalArchivist
- Geographic structure generation with AI

## Directory Structure

- Service implementation: `services/worldgenerator/`
- Command entry point: `cmd/world-generator/`
- Internal packages: `internal/eventbus/`, `internal/oracle/`

## Service-Specific Conventions

- Uses UniverseGenesisOracle to request world schemas
- Generates entity types: world, region, zone, city, water_body
- Integrates with EntityManager for entity creation
- Uses SemanticMemory for contextual generation
- Supports manual and event-driven initialization
- Entity type for worlds: `entity_type: world`
- Schema storage path: `schemas/{type}/{name}/v{version}.json`
- Publishes geographic events for CityGovernor integration