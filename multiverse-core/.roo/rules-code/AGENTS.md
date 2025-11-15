# Project Coding Rules (Non-Obvious Only)

- Always use the custom Entity structure with history tracking and path-based payload access from internal/entity package
- World generator uses Ascension Oracle (Qwen3) for generating world details and schemas - this is mandatory for proper world generation
- Narrative orchestrator uses Semantic Memory Builder (ChromaDB + Neo4j) for RAG context - this is required for proper narrative generation
- Ontological Archivist stores schemas in MinIO with versioned JSON files - this is the only correct way to store schemas
- BanOfWorld service uses "resonance" metrics for world integrity monitoring - these metrics are critical for world integrity
- CultivationModule generates "Dao portraits" from player history for ascension events - this is required for proper ascension handling
- All services use structured logging with consistent format - use the provided logging patterns
- Services use context for cancellation and timeouts - all methods must accept context parameter
- Event bus uses Kafka with LeastBytes balancer for load distribution - this is the only supported balancer
- Services implement graceful shutdown with HTTP server shutdown and resource cleanup - all services must handle shutdown properly
- Entity manager supports both state changes and full entity snapshots - both approaches are valid and must be used appropriately
- All services must be able to recover from snapshots and replay events - this is a core requirement
- Services must handle entity travel between worlds through snapshot management - this is required for world transitions
- All services must be able to handle concurrent access to shared resources - use proper synchronization
- Services use specific event types for different operations (entity.created, entity.updated, world.generated, etc.) - these are the only valid event types
- World generation requires specific seed-based prompts for Qwen3 - these prompts are critical for proper generation
- All services must be built with CGO_ENABLED=0 for cross-platform compatibility - this is required for all builds
- Services use specific environment variables for configuration (MINIO_ENDPOINT, ORACLE_URL, SEMANTIC_MEMORY_URL, etc.) - these are the only valid environment variables