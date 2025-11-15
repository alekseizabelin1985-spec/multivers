// Package worldgenerator implements the WorldGenerator service.
package worldgenerator

import (
	"context"

	"multiverse-core/internal/eventbus"
)

// Service manages the WorldGenerator lifecycle.
type Service struct {
	bus       *eventbus.EventBus
	generator *WorldGenerator
}

// NewService creates a new WorldGenerator service.
func NewService(bus *eventbus.EventBus) *Service {
	return &Service{
		bus:       bus,
		generator: NewWorldGenerator(bus),
	}
}

// Run starts the service and blocks until context is cancelled.
func (s *Service) Run(ctx context.Context) error {
	s.bus.Subscribe(ctx, eventbus.TopicSystemEvents, "world-generator-group", s.generator.HandleEvent)
	<-ctx.Done()
	return ctx.Err()
}
