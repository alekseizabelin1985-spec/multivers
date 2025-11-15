# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Project Overview

This is a multiverse game engine built with event-driven architecture. The system consists of multiple specialized services that communicate through an event bus (Redpanda/Kafka).

Each service has its own AGENTS.md file in its respective directory under `services/{service-name}/AGENTS.md` with specific guidance for working with that service.

## Build/lint/test commands

- Build all services: `make build`
- Build specific service: `make build-service SERVICE=entity-manager`
- Run specific service: `make run SERVICE=entity-manager`
- Start all services: `make up`
- Stop all services: `make down`
- Clean build artifacts: `make clean`
- Show logs for all services: `make logs`
- Show logs for specific service: `make logs-service SERVICE=entity-manager`

## Code style guidelines

- All services are written in Go 1.25
- Use JSON Schema Draft 7 for entity payload validation
- Follow event-driven architecture with Kafka/Redpanda as message broker
- Services communicate through events in the event bus (topics: player_events, world_events, game_events, system_events, scope_management, narrative_output)
- Entity management uses MinIO for storage with bucket naming pattern: `entities-{world_id}`
- All services are stateful and support recovery via snapshots and event replay
- Use entity paths with dot notation for nested payload access (e.g., `payload.health.current`)
- All services must be built with CGO disabled for cross-platform compatibility
- Use UUIDs for event IDs and entity IDs
- All services must be containerized with Docker using multi-stage builds

## Custom utilities and patterns

- Entity manager uses a custom entity structure with history tracking and path-based payload access
- World generator uses Ascension Oracle (Qwen3) for generating world details and schemas
- Narrative orchestrator uses Semantic Memory Builder (ChromaDB + Neo4j) for RAG context
- Ontological Archivist stores schemas in MinIO with versioned JSON files
- BanOfWorld service uses "resonance" metrics for world integrity monitoring
- CultivationModule generates "Dao portraits" from player history for ascension events
- All services use structured logging with consistent format
- Services use context for cancellation and timeouts
- Event bus uses Kafka with LeastBytes balancer for load distribution
- Services implement graceful shutdown with HTTP server shutdown and resource cleanup

## Non-standard directory structures

- Services are organized in `services/` directory with each service in its own subdirectory
- Commands are in `cmd/` directory with each command in its own subdirectory
- Internal packages are in `internal/` directory
- Documentation is in `Docs/` directory
- Docker configuration in `build/` directory
- Fake dependencies for CGO issues in `fake_deps/` directory

## Project-specific conventions

- All services must be built with CGO_ENABLED=0 for cross-platform compatibility
- Services use specific naming conventions for topics: `player_events`, `world_events`, `game_events`, `system_events`, `scope_management`, `narrative_output`
- World entities are stored in MinIO buckets named `entities-{world_id}`
- Global entities are stored in `entities-global` bucket
- Schema versions are stored in MinIO with path pattern: `schemas/{type}/{name}/v{version}.json`
- All services must be run with Docker Compose for proper environment setup
- Services communicate through event bus, not direct API calls
- Entity manager supports both state changes and full entity snapshots
- All services must be able to recover from snapshots and replay events
- Services use specific environment variables for configuration (MINIO_ENDPOINT, ORACLE_URL, SEMANTIC_MEMORY_URL, etc.)
- World generation requires specific seed-based prompts for Qwen3
- Services must handle entity travel between worlds through snapshot management
- All services must be able to handle concurrent access to shared resources
- Services use specific event types for different operations (entity.created, entity.updated, world.generated, etc.)
- Semantic Memory service stores all events for context and replay

## Service-Specific Guidance

Each service has detailed guidance in its own AGENTS.md file:

- [EntityManager](services/entitymanager/AGENTS.md) - Manages game entities and their history
- [NarrativeOrchestrator](services/narrativeorchestrator/AGENTS.md) - Generates dynamic narratives
- [WorldGenerator](services/worldgenerator/AGENTS.md) - Creates game worlds
- [BanOfWorld](services/banofworld/AGENTS.md) - Monitors world integrity
- [CityGovernor](services/citygovernor/AGENTS.md) - Manages city structures
- [CultivationModule](services/cultivationmodule/AGENTS.md) - Handles player cultivation
- [RealityMonitor](services/realitymonitor/AGENTS.md) - Monitors game reality state
- [PlanManager](services/planmanager/AGENTS.md) - Manages player plans
- [SemanticMemory](services/semanticmemory/AGENTS.md) - Provides semantic context
- [OntologicalArchivist](services/ontologicalarchivist/AGENTS.md) - Stores schemas
- [UniverseGenesisOracle](services/universegenesis/AGENTS.md) - Generates universes
- [GameService](services/gameservice/AGENTS.md) - Handles player interactions
- [AscensionOracle](services/ascensionoracle/AGENTS.md) - Provides AI generation