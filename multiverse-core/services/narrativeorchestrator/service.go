// services/narrativeorchestrator/service.go
package narrativeorchestrator

import (
	"context"
	"log"

	"multiverse-core/internal/eventbus"
)

type Config struct {
	KafkaBrokers []string
}

type Service struct {
	orchestrator *NarrativeOrchestrator
	bus          *eventbus.EventBus
}

func NewService(cfg Config) (*Service, error) {
	bus := eventbus.NewEventBus(cfg.KafkaBrokers)
	orchestrator := NewNarrativeOrchestrator(bus)

	return &Service{
		orchestrator: orchestrator,
		bus:          bus,
	}, nil
}

func (s *Service) Start(ctx context.Context) {
	log.Println("NarrativeOrchestrator started. Listening to system_events, world_events, game_events...")

	// System events for GM lifecycle
	go s.bus.Subscribe(ctx, eventbus.TopicSystemEvents, "narrative-orchestrator-group", func(ev eventbus.Event) {
		switch ev.EventType {
		case "gm.created":
			s.orchestrator.CreateGM(ev)
		case "gm.deleted":
			s.orchestrator.DeleteGM(ev)
		case "gm.merged":
			s.orchestrator.MergeGM(ev)
		case "gm.split":
			s.orchestrator.SplitGM(ev)
		}
	})

	// Game events for active GMs
	go s.bus.Subscribe(ctx, eventbus.TopicWorldEvents, "narrative-world-group", s.orchestrator.HandleGameEvent)
	go s.bus.Subscribe(ctx, eventbus.TopicGameEvents, "narrative-game-group", s.orchestrator.HandleGameEvent)
}

func (s *Service) Stop() {
	s.bus.Close()
}
