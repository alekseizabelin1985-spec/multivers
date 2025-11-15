# AGENTS.md for PlanManager

This file provides guidance to AI assistants when working with the PlanManager service.

## Service Overview

PlanManager manages strategic and tactical plans for players and NPCs.

### Key Responsibilities
- Creating and managing strategic plans
- Aggregating and executing tactical tasks
- Supporting multi-player planning
- Generating plans based on game world context

### Event Integration
- Subscribes to: `player_events` and `world_events` topics
- Publishes: plan status events
- Event types: `plan.created`, `plan.updated`, `plan.deleted`, `plan.executed`, `plan.completed`, `plan.failed`

## Build/Run Commands

- Build: `make build-service SERVICE=plan-manager`
- Run: `make run SERVICE=plan-manager`
- Logs: `make logs-service SERVICE=plan-manager`

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
- Plan state management in memory
- Task execution tracking
- Context-based plan generation

## Directory Structure

- Service implementation: `services/planmanager/`
- Command entry point: `cmd/plan-manager/
- Internal packages: `internal/eventbus/`

## Service-Specific Conventions

- Stores plan data in memory
- Tracks task execution status
- Maintains plan change history
- Integrates with EntityManager for entity data
- Works with WorldGenerator for world context
- Communicates with SemanticMemory for planning context
- Event group subscriptions: `plan-manager-group`, `plan-world-group`
- Supports complex, multi-step plans
- Event-driven plan lifecycle management