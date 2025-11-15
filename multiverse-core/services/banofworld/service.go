// Package banofworld implements the BanOfWorld service.
package banofworld

import (
	"context"

	"multiverse-core/internal/eventbus"
)

// Service manages the BanOfWorld lifecycle.
type Service struct {
	bus *eventbus.EventBus
	ban *BanOfWorld
}

// NewService creates a new BanOfWorld service.
func NewService(bus *eventbus.EventBus) *Service {
	return &Service{
		bus: bus,
		ban: NewBanOfWorld(bus),
	}
}

// Run starts the service and blocks until context is cancelled.
func (s *Service) Run(ctx context.Context) error {
	// Subscribe to player_events for integrity checks
	s.bus.Subscribe(ctx, eventbus.TopicPlayerEvents, "ban-of-world-group", s.ban.HandlePlayerEvent)
	<-ctx.Done()
	return ctx.Err()
}
