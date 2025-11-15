# AGENTS.md for CultivationModule

This file provides guidance to AI assistants when working with the CultivationModule service.

## Service Overview

CultivationModule generates "Dao portraits" of players for ascension events.

### Key Responsibilities
- Generating Dao portraits based on player history
- Preparing data for ascension events
- Tracking player progress in cultivation practice
- Supporting character development

### Event Integration
- Subscribes to: `player_events` topic with player action event types
- Publishes: Dao portrait generation results
- Event types: player skill use, ascension trigger events

## Build/Run Commands

- Build: `make build-service SERVICE=cultivation-module`
- Run: `make run SERVICE=cultivation-module`
- Logs: `make logs-service SERVICE=cultivation-module`

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
- Player history analysis
- Dao portrait generation
- Integration with ascension mechanics

## Directory Structure

- Service implementation: `services/cultivationmodule/`
- Command entry point: `cmd/cultivation-module/
- Internal packages: `internal/eventbus/`

## Service-Specific Conventions

- Tracks player cultivation history
- Generates personalized Dao portraits
- Supports hierarchical modules at each plan level
- Integrates with NarrativeOrchestrator for context
- Works with BanOfWorld for integrity monitoring
- Event types: `dao.portrait.updated`, `ascension.trial.started`
- Stateless processing with real-time event handling
- Player-focused event processing