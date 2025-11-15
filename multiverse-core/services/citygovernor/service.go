// Package citygovernor implements the CityGovernor service.
package citygovernor

import (
	"context"

	"multiverse-core/internal/eventbus"
)

// Service manages the CityGovernor lifecycle.
type Service struct {
	bus      *eventbus.EventBus
	governor *CityGovernor
}

// NewService creates a new CityGovernor service.
func NewService(bus *eventbus.EventBus) *Service {
	return &Service{
		bus:      bus,
		governor: NewCityGovernor(bus),
	}
}

// Run starts the service and blocks until context is cancelled.
func (s *Service) Run(ctx context.Context) error {
	// Subscribe to game_events and world_events for city management
	topics := []string{
		eventbus.TopicGameEvents,
		eventbus.TopicWorldEvents,
	}

	for _, topic := range topics {
		topic := topic // Capture range variable
		go s.bus.Subscribe(ctx, topic, "city-governor-group", s.governor.HandleEvent)
	}

	<-ctx.Done()
	return ctx.Err()
}
