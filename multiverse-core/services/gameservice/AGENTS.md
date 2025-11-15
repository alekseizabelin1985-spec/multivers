# AGENTS.md for GameService

This file provides guidance to AI assistants when working with the GameService.

## Service Overview

GameService handles player interactions and game mechanics in the multiverse.

### Key Responsibilities
- Managing player connections and interactions
- Handling game mechanics and rules
- Coordinating with other services for game state
- Providing WebSocket and HTTP interfaces

### Event Integration
- Subscribes to: Multiple event topics for game state
- Publishes: Player action events
- Event types: Player actions, game state changes

## Build/Run Commands

- Build: `make build-service SERVICE=gameservice`
- Run: `make run SERVICE=gameservice`
- Logs: `make logs-service SERVICE=gameservice`

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
- WebSocket communication with players
- HTTP API for game operations
- Cache management for performance
- Integration with EntityManager for entity operations

## Directory Structure

- Service implementation: `services/gameservice/`
- Command entry point: `cmd/gameservice/
- Internal packages: `internal/eventbus/`

## Service-Specific Conventions

- Provides HTTP API on port 8088 (default)
- WebSocket connections for real-time player interaction
- Cache management with 5-minute TTL (default)
- Integrates with EntityManager for entity operations
- Player service coordination
- MinIO client for storage operations
- HTTP server with graceful shutdown
- Concurrent player connection handling
- Event-driven player action processing