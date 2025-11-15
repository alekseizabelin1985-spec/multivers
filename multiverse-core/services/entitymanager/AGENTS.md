# AGENTS.md for EntityManager

This file provides guidance to AI assistants when working with the EntityManager service.

## Service Overview

EntityManager manages hierarchical entities in the game world, ensuring their state and history.

### Key Responsibilities
- Managing entity states and history
- Supporting entity snapshot/replay for recovery
- Isolating data by world (entities-{world_id} buckets)
- Handling entity events (created, updated, deleted)

### Event Integration
- Subscribes to: `entity_events` topic with types `entity.created`, `entity.updated`, `entity.deleted`
- Publishes: entity operation confirmations

## Build/Run Commands

- Build: `make build-service SERVICE=entity-manager`
- Run: `make run SERVICE=entity-manager`
- Logs: `make logs-service SERVICE=entity-manager`

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
- MinIO storage for entity snapshots
- Snapshot/replay recovery mechanism
- World-based data isolation

## Directory Structure

- Service implementation: `services/entitymanager/`
- Command entry point: `cmd/entity-manager/`
- Internal packages: `internal/entity/`, `internal/eventbus/`, `internal/storage/`

## Service-Specific Conventions

- Entities stored in MinIO buckets named `entities-{world_id}`
- Supports both state changes and full entity snapshots
- Must handle entity travel between worlds through snapshot management
- Uses specific event types: `entity.created`, `entity.updated`, `entity.history.appended`