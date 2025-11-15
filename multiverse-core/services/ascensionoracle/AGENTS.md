# AGENTS.md for AscensionOracle

This file provides guidance to AI assistants when working with the AscensionOracle service.

## Service Overview

AscensionOracle is a specialized AI service ensuring philosophical integrity and generativity in responses.

### Key Responsibilities
- Generating high-quality AI responses with philosophical depth
- Supporting "generativity vs scripts" approach
- Ensuring semantic integrity in content
- Providing context for other AI services

### Event Integration
- Subscribes to: `system_events` topic with type `oracle.request`
- Publishes: AI-generated responses
- Event types: oracle request/response events

## Build/Run Commands

- Build: `make build-service SERVICE=ascension-oracle`
- Run: `make run SERVICE=ascension-oracle`
- Logs: `make logs-service SERVICE=ascension-oracle`

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
- Qwen3 integration for generation
- Philosophical integrity enforcement

## Directory Structure

- Service implementation: `services/ascensionoracle/` (if exists)
- Command entry point: `cmd/ascension-oracle/` (if exists)
- Internal packages: `internal/eventbus/`, `internal/oracle/`

## Service-Specific Conventions

- Stateless between requests
- Uses Qwen3 for generation
- Ensures philosophical integrity and semantic depth
- Integrates with UniverseGenesisOracle for context
- Works with NarrativeOrchestrator for storytelling
- Communicates with SemanticMemory for context
- Supports WorldGenerator for world context
- Generativity over scripted responses
- HTTP-based AI service integration
- Event-driven request processing