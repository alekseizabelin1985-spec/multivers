# AGENTS.md for CityGovernor

This file provides guidance to AI assistants when working with the CityGovernor service.

## Service Overview

CityGovernor manages city structures and ensures their functioning in the game world.

### Key Responsibilities
- Managing city entities and structures
- Ensuring city functionality
- Supporting interaction between city elements
- Coordinating actions within cities

### Event Integration
- Subscribes to: `world_events` topic with city-related event types
- Publishes: city operation results
- Event types: player enter/exit city, trade events, crime events

## Build/Run Commands

- Build: `make build-service SERVICE=city-governor`
- Run: `make run SERVICE=city-governor`
- Logs: `make logs-service SERVICE=city-governor`

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
- Real-time city state management
- Coordination between city elements
- Integration with game mechanics

## Directory Structure

- Service implementation: `services/citygovernor/`
- Command entry point: `cmd/city-governor/
- Internal packages: `internal/eventbus/`

## Service-Specific Conventions

- Manages city metrics: reputation, crime_rate, active_quests
- Publishes events: `quest.issued`, `market.price_changed`, `festival.started`
- Responds to group composition (rich/poor players)
- Can generate unique quests through AI
- Integrates with WorldGenerator for city creation
- Works with PlanManager for city planning
- Uses real-time event processing