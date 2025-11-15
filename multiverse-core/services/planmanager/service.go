// Package planmanager implements the PlanManager service.
package planmanager

import (
	"context"

	"multiverse-core/internal/eventbus"
)

// Service manages the PlanManager lifecycle.
type Service struct {
	bus     *eventbus.EventBus
	manager *PlanManager
}

// NewService creates a new PlanManager service.
func NewService(bus *eventbus.EventBus) *Service {
	return &Service{
		bus:     bus,
		manager: NewPlanManager(bus),
	}
}

// Run starts the service and blocks until context is cancelled.
func (s *Service) Run(ctx context.Context) error {
	// Subscribe to world_events for ascension and convergence events
	s.bus.Subscribe(ctx, eventbus.TopicWorldEvents, "plan-manager-group", s.manager.HandleWorldEvent)

	// Also subscribe to system_events for world generation
	s.bus.Subscribe(ctx, eventbus.TopicSystemEvents, "plan-manager-group", s.manager.HandleWorldEvent)

	<-ctx.Done()
	return ctx.Err()
}
