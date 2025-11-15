// Package planmanager handles plan hierarchy and ascension routing.
package planmanager

import (
	"context"
	"log"
	"time"

	"multiverse-core/internal/eventbus"

	"github.com/google/uuid"
)

// PlanManager manages the hierarchy of plans and ascension routing.
type PlanManager struct {
	bus *eventbus.EventBus
}

// NewPlanManager creates a new PlanManager.
func NewPlanManager(bus *eventbus.EventBus) *PlanManager {
	return &PlanManager{bus: bus}
}

// HandleWorldEvent processes world events for plan management.
func (pm *PlanManager) HandleWorldEvent(ev eventbus.Event) {
	switch ev.EventType {
	case "ascension.attempt":
		pm.routeAscension(ev)
	case "plan.convergence.requested":
		pm.activateConvergenceZone(ev)
	case "world.generated":
		pm.initializeWorldPlan(ev)
	}
}

// routeAscension routes an ascension attempt to the appropriate plan.
func (pm *PlanManager) routeAscension(ev eventbus.Event) {
	currentPlan, _ := ev.Payload["current_plan"].(float64)
	playerID, _ := ev.Payload["player_id"].(string)

	if playerID == "" {
		log.Printf("Ascension attempt missing player_id")
		return
	}

	targetPlan := int(currentPlan + 1)
	targetWorld := pm.getTargetWorldForPlan(targetPlan, ev.WorldID)

	routeEvent := eventbus.Event{
		EventID:   "route-" + uuid.New().String()[:8],
		EventType: "ascension.routed",
		Source:    "plan-manager",
		WorldID:   ev.WorldID,
		Payload: map[string]interface{}{
			"player_id":    playerID,
			"from_plan":    currentPlan,
			"to_plan":      targetPlan,
			"target_world": targetWorld,
			"ritual_id":    ev.Payload["ritual_id"],
		},
		Timestamp: time.Now(),
	}

	pm.bus.Publish(context.Background(), eventbus.TopicSystemEvents, routeEvent)
	log.Printf("Ascension for %s routed from Plan %d to Plan %d (world: %s)",
		playerID, int(currentPlan), targetPlan, targetWorld)
}

// activateConvergenceZone activates a convergence zone for plan merging.
func (pm *PlanManager) activateConvergenceZone(ev eventbus.Event) {
	planLevel, _ := ev.Payload["plan_level"].(float64)
	worldID, _ := ev.Payload["world_id"].(string)

	if worldID == "" {
		log.Printf("Convergence request missing world_id")
		return
	}

	zoneEvent := eventbus.Event{
		EventID:   "zone-" + uuid.New().String()[:8],
		EventType: "plan.convergence.activated",
		Source:    "plan-manager",
		WorldID:   worldID,
		Payload: map[string]interface{}{
			"zone_id":        "convergence-zone-" + worldID,
			"plan_level":     planLevel,
			"duration_hours": 24,
			"participants":   ev.Payload["participants"],
		},
		Timestamp: time.Now(),
	}

	pm.bus.Publish(context.Background(), eventbus.TopicSystemEvents, zoneEvent)
	log.Printf("Convergence zone activated for %s at Plan %d", worldID, int(planLevel))
}

// initializeWorldPlan initializes the plan level for a newly generated world.
func (pm *PlanManager) initializeWorldPlan(ev eventbus.Event) {
	worldID, _ := ev.Payload["world_id"].(string)
	if worldID == "" {
		return
	}

	// Determine plan level based on world seed or constraints
	planLevel := 0 // Default to Plan 0 (base worlds)

	if constraints, ok := ev.Payload["constraints"].([]interface{}); ok {
		for _, constraint := range constraints {
			if constraint == "high_plan" {
				planLevel = 1
				break
			}
		}
	}

	initEvent := eventbus.Event{
		EventID:   "plan-init-" + uuid.New().String()[:8],
		EventType: "plan.initialized",
		Source:    "plan-manager",
		WorldID:   worldID,
		Payload: map[string]interface{}{
			"world_id":   worldID,
			"plan_level": planLevel,
		},
		Timestamp: time.Now(),
	}

	pm.bus.Publish(context.Background(), eventbus.TopicSystemEvents, initEvent)
	log.Printf("World %s initialized at Plan %d", worldID, planLevel)
}

// getTargetWorldForPlan determines the target world for ascension to a plan.
func (pm *PlanManager) getTargetWorldForPlan(plan int, currentWorld string) string {
	switch plan {
	case 1:
		return "convergence-zone-1"
	case 2:
		return "abstract-realm"
	default:
		if plan > 2 {
			return "plan-omega"
		}
		return currentWorld // Fallback
	}
}
