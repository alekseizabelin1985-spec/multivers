# AGENTS.md for BanOfWorld

This file provides guidance to AI assistants when working with the BanOfWorld (World Ban) service.

## Service Overview

BanOfWorld monitors world integrity and applies measures when violations occur.

### Key Responsibilities
- Monitoring world "resonance" for integrity
- Detecting and responding to integrity violations
- Applying corrective measures
- Supporting ontological integrity

### Event Integration
- Subscribes to: `system_events` topic with type `world.integrity.check`
- Publishes: integrity check results
- Event types: `world.integrity.check`, anomaly detection events

## Build/Run Commands

- Build: `make build-service SERVICE=ban-of-world`
- Run: `make run SERVICE=ban-of-world`
- Logs: `make logs-service SERVICE=ban-of-world`

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
- Real-time world state monitoring
- Resonance metrics evaluation
- Integration with AscensionOracle for anomaly resolution

## Directory Structure

- Service implementation: `services/banofworld/`
- Command entry point: `cmd/ban-of-world/
- Internal packages: `internal/eventbus/`

## Service-Specific Conventions

- Monitors metrics: `spatial_integrity`, `karma_entropy`, `core_resonance`
- Uses "resonance" metrics for integrity assessment
- Triggers AscensionOracle as Oracle for anomalies
- Generates mythological consequences rather than penalties
- Stateless monitoring with real-time event processing
- Configurable resonance threshold (default: 0.8)