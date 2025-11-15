# AGENTS.md for RealityMonitor

This file provides guidance to AI assistants when working with the RealityMonitor service.

## Service Overview

RealityMonitor tracks and analyzes the state of the game reality.

### Key Responsibilities
- Monitoring game reality state
- Detecting anomalies and inconsistencies
- Analyzing world changes
- Providing data for corrections

### Event Integration
- Subscribes to: `system_events` topic with type `reality.check`
- Publishes: reality analysis results
- Event types: reality check, anomaly detection events

## Build/Run Commands

- Build: `make build-service SERVICE=reality-monitor`
- Run: `make run SERVICE=reality-monitor`
- Logs: `make logs-service SERVICE=reality-monitor`

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
- Real-time reality state monitoring
- Anomaly detection algorithms
- Integration with corrective systems

## Directory Structure

- Service implementation: `services/realitymonitor/`
- Command entry point: `cmd/reality-monitor/
- Internal packages: `internal/eventbus/`

## Service-Specific Conventions

- Aggregates metrics from all worlds
- Uses various metrics for analysis
- Triggers corrective actions when needed
- Integrates with BanOfWorld for integrity monitoring
- Works with WorldGenerator for world information
- Communicates with SemanticMemory for context
- Stateless monitoring with real-time processing
- Event-driven anomaly detection