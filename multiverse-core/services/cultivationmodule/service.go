// Package cultivationmodule implements the CultivationModule service.
package cultivationmodule

import (
	"context"

	"multiverse-core/internal/eventbus"
)

// Service manages the CultivationModule lifecycle.
type Service struct {
	bus         *eventbus.EventBus
	cultivation *CultivationModule
}

// NewService creates a new CultivationModule service.
func NewService(bus *eventbus.EventBus) *Service {
	return &Service{
		bus:         bus,
		cultivation: NewCultivationModule(bus),
	}
}

// Run starts the service and blocks until context is cancelled.
func (s *Service) Run(ctx context.Context) error {
	// Subscribe to relevant event topics
	topics := []string{
		eventbus.TopicPlayerEvents,
		eventbus.TopicWorldEvents,
		eventbus.TopicSystemEvents,
	}

	for _, topic := range topics {
		topic := topic // Capture range variable
		go s.bus.Subscribe(ctx, topic, "cultivation-module-group", s.cultivation.HandleEvent)
	}

	<-ctx.Done()
	return ctx.Err()
}
