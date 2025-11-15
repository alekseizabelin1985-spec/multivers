// Package citygovernor manages city life, NPCs, quests, and economy.
package citygovernor

import (
	"context"
	"log"
	"time"

	"multiverse-core/internal/eventbus"

	"github.com/google/uuid"
)

// CityGovernor manages city-related logic.
type CityGovernor struct {
	bus *eventbus.EventBus
}

// NewCityGovernor creates a new CityGovernor.
func NewCityGovernor(bus *eventbus.EventBus) *CityGovernor {
	return &CityGovernor{bus: bus}
}

// HandleEvent processes events for city management.
func (cg *CityGovernor) HandleEvent(ev eventbus.Event) {
	if ev.ScopeID == nil {
		return // Not a city-scoped event
	}

	switch ev.EventType {
	case "player.entered":
		cg.handlePlayerEntry(ev)
	case "violation.detected":
		cg.handleViolation(ev)
	case "quest.completed":
		cg.handleQuestCompletion(ev)
	case "city.reputation.changed":
		cg.handleReputationChange(ev)
	case "npc.interaction":
		cg.handleNPCInteraction(ev)
	}
}

// handlePlayerEntry handles player entry into a city.
func (cg *CityGovernor) handlePlayerEntry(ev eventbus.Event) {
	playerID, _ := ev.Payload["player_id"].(string)
	cityID := *ev.ScopeID

	if playerID == "" {
		log.Printf("Player entry missing player_id")
		return
	}

	// Generate welcome quest if player is new
	if cg.isNewPlayerInCity(playerID, cityID) {
		cg.generateWelcomeQuest(ev)
	}

	// Update city population
	cg.updateCityPopulation(cityID, 1)

	// Notify NPCs of new arrival
	npcEvent := eventbus.Event{
		EventID:   "npc-notify-" + uuid.New().String()[:8],
		EventType: "citizen.event",
		Source:    "city-governor",
		WorldID:   ev.WorldID,
		ScopeID:   &cityID,
		Payload: map[string]interface{}{
			"event_type": "new_arrival",
			"player_id":  playerID,
			"city_id":    cityID,
		},
		Timestamp: time.Now(),
	}
	cg.bus.Publish(context.Background(), eventbus.TopicGameEvents, npcEvent)

	log.Printf("Player %s entered city %s", playerID, cityID)
}

// handleViolation handles world integrity violations in the city.
func (cg *CityGovernor) handleViolation(ev eventbus.Event) {
	cityID := *ev.ScopeID
	playerID, _ := ev.Payload["player_id"].(string)
	violationType, _ := ev.Payload["violation_type"].(string)

	if playerID == "" {
		return
	}

	// Apply city-specific consequences
	consequence := cg.getCityConsequence(violationType, cityID)

	consequenceEvent := eventbus.Event{
		EventID:   "city-conseq-" + uuid.New().String()[:8],
		EventType: "city.violation.consequence",
		Source:    "city-governor",
		WorldID:   ev.WorldID,
		ScopeID:   &cityID,
		Payload: map[string]interface{}{
			"player_id":      playerID,
			"violation_type": violationType,
			"consequence":    consequence,
			"city_id":        cityID,
		},
		Timestamp: time.Now(),
	}
	cg.bus.Publish(context.Background(), eventbus.TopicGameEvents, consequenceEvent)

	// Update city reputation
	cg.updateCityReputation(cityID, -10) // Reputation decreases on violations

	log.Printf("Applied consequence %s for violation %s in city %s", consequence, violationType, cityID)
}

// handleQuestCompletion handles quest completion events.
func (cg *CityGovernor) handleQuestCompletion(ev eventbus.Event) {
	playerID, _ := ev.Payload["player_id"].(string)
	questID, _ := ev.Payload["quest_id"].(string)
	cityID := *ev.ScopeID

	if playerID == "" || questID == "" {
		return
	}

	// Generate reward
	reward := cg.generateQuestReward(questID, cityID)

	rewardEvent := eventbus.Event{
		EventID:   "quest-reward-" + uuid.New().String()[:8],
		EventType: "quest.reward.granted",
		Source:    "city-governor",
		WorldID:   ev.WorldID,
		ScopeID:   &cityID,
		Payload: map[string]interface{}{
			"player_id": playerID,
			"quest_id":  questID,
			"reward":    reward,
			"city_id":   cityID,
		},
		Timestamp: time.Now(),
	}
	cg.bus.Publish(context.Background(), eventbus.TopicGameEvents, rewardEvent)

	// Update reputation based on quest type
	questType, _ := ev.Payload["quest_type"].(string)
	reputationChange := cg.getReputationChangeForQuest(questType)
	cg.updateCityReputation(cityID, reputationChange)

	// Generate new quest
	cg.generateNewQuest(ev)

	log.Printf("Granted reward for quest %s to player %s in city %s", questID, playerID, cityID)
}

// handleReputationChange handles reputation changes.
func (cg *CityGovernor) handleReputationChange(ev eventbus.Event) {
	cityID := *ev.ScopeID
	change, _ := ev.Payload["change"].(float64)

	newReputation := cg.getCurrentReputation(cityID) + int(change)

	// Apply reputation effects
	cg.applyReputationEffects(cityID, newReputation)

	log.Printf("City %s reputation changed to %d", cityID, newReputation)
}

// handleNPCInteraction handles NPC interaction events.
func (cg *CityGovernor) handleNPCInteraction(ev eventbus.Event) {
	playerID, _ := ev.Payload["player_id"].(string)
	npcID, _ := ev.Payload["npc_id"].(string)
	interactionType, _ := ev.Payload["interaction_type"].(string)
	cityID := *ev.ScopeID

	if playerID == "" || npcID == "" {
		return
	}

	// Generate interaction response
	response := cg.generateNPCResponse(npcID, interactionType, cityID)

	responseEvent := eventbus.Event{
		EventID:   "npc-response-" + uuid.New().String()[:8],
		EventType: "npc.response.generated",
		Source:    "city-governor",
		WorldID:   ev.WorldID,
		ScopeID:   &cityID,
		Payload: map[string]interface{}{
			"player_id":        playerID,
			"npc_id":           npcID,
			"interaction_type": interactionType,
			"response":         response,
			"city_id":          cityID,
		},
		Timestamp: time.Now(),
	}
	cg.bus.Publish(context.Background(), eventbus.TopicGameEvents, responseEvent)

	log.Printf("Generated NPC response for %s -> %s in city %s", playerID, npcID, cityID)
}

// generateWelcomeQuest generates a welcome quest for new players.
func (cg *CityGovernor) generateWelcomeQuest(ev eventbus.Event) {
	playerID, _ := ev.Payload["player_id"].(string)
	cityID := *ev.ScopeID

	questID := "welcome-" + uuid.New().String()[:8]

	questEvent := eventbus.Event{
		EventID:   "quest-welcome-" + uuid.New().String()[:8],
		EventType: "quest.assigned",
		Source:    "city-governor",
		WorldID:   ev.WorldID,
		ScopeID:   &cityID,
		Payload: map[string]interface{}{
			"quest_id":    questID,
			"player_id":   playerID,
			"city_id":     cityID,
			"title":       "Добро пожаловать в " + cg.getCityName(cityID),
			"description": "Старейшина просит вас принести ему Слёзы Памяти из ближайшего леса.",
			"reward":      "50 золотых и репутация +10",
			"quest_type":  "welcome",
		},
		Timestamp: time.Now(),
	}
	cg.bus.Publish(context.Background(), eventbus.TopicGameEvents, questEvent)
}

// generateNewQuest generates a new quest after completion.
func (cg *CityGovernor) generateNewQuest(ev eventbus.Event) {
	playerID, _ := ev.Payload["player_id"].(string)
	cityID := *ev.ScopeID

	// Determine quest type based on player history and city state
	questType := cg.determineNextQuestType(playerID, cityID)

	questID := "quest-" + uuid.New().String()[:8]

	questEvent := eventbus.Event{
		EventID:   "quest-new-" + uuid.New().String()[:8],
		EventType: "quest.assigned",
		Source:    "city-governor",
		WorldID:   ev.WorldID,
		ScopeID:   &cityID,
		Payload: map[string]interface{}{
			"quest_id":    questID,
			"player_id":   playerID,
			"city_id":     cityID,
			"title":       cg.generateQuestTitle(questType, cityID),
			"description": cg.generateQuestDescription(questType, cityID),
			"reward":      cg.generateQuestReward(questID, cityID),
			"quest_type":  questType,
		},
		Timestamp: time.Now(),
	}
	cg.bus.Publish(context.Background(), eventbus.TopicGameEvents, questEvent)
}

// Helper methods (simplified implementations)

func (cg *CityGovernor) isNewPlayerInCity(playerID, cityID string) bool {
	// In a real implementation, this would check player history
	return true // Simplified for example
}

func (cg *CityGovernor) updateCityPopulation(cityID string, delta int) {
	// Publish population update event
	popEvent := eventbus.Event{
		EventID:   "pop-update-" + uuid.New().String()[:8],
		EventType: "city.population.changed",
		Source:    "city-governor",
		WorldID:   "global", // Population affects world state
		Payload: map[string]interface{}{
			"city_id": cityID,
			"delta":   delta,
		},
		Timestamp: time.Now(),
	}
	cg.bus.Publish(context.Background(), eventbus.TopicGameEvents, popEvent)
}

func (cg *CityGovernor) getCityConsequence(violationType, cityID string) string {
	// City-specific consequences
	switch cityID {
	case "city-ashes":
		return "imprisonment" // Harsh consequences in City of Ashes
	case "city-archives":
		return "memory_wipe" // Memory-based consequences in City of Archives
	default:
		return "fine" // Default consequence
	}
}

func (cg *CityGovernor) updateCityReputation(cityID string, delta int) {
	repEvent := eventbus.Event{
		EventID:   "rep-update-" + uuid.New().String()[:8],
		EventType: "city.reputation.changed",
		Source:    "city-governor",
		WorldID:   "global",
		Payload: map[string]interface{}{
			"city_id": cityID,
			"change":  delta,
		},
		Timestamp: time.Now(),
	}
	cg.bus.Publish(context.Background(), eventbus.TopicGameEvents, repEvent)
}

func (cg *CityGovernor) generateQuestReward(questID, cityID string) string {
	// Generate reward based on quest and city
	return "100 золотых и уникальный предмет"
}

func (cg *CityGovernor) getReputationChangeForQuest(questType string) int {
	switch questType {
	case "help_citizen":
		return 5
	case "defeat_monster":
		return 10
	case "violation":
		return -15
	default:
		return 0
	}
}

func (cg *CityGovernor) getCurrentReputation(cityID string) int {
	// In real implementation, this would query city state
	return 50 // Default reputation
}

func (cg *CityGovernor) applyReputationEffects(cityID string, reputation int) {
	// Apply effects based on reputation level
	var effect string
	if reputation > 75 {
		effect = "prosperity" // City thrives
	} else if reputation < 25 {
		effect = "decay" // City deteriorates
	} else {
		effect = "stability" // Normal state
	}

	effectEvent := eventbus.Event{
		EventID:   "rep-effect-" + uuid.New().String()[:8],
		EventType: "city.reputation.effect",
		Source:    "city-governor",
		WorldID:   "global",
		Payload: map[string]interface{}{
			"city_id": cityID,
			"effect":  effect,
			"level":   reputation,
		},
		Timestamp: time.Now(),
	}
	cg.bus.Publish(context.Background(), eventbus.TopicGameEvents, effectEvent)
}

func (cg *CityGovernor) generateNPCResponse(npcID, interactionType, cityID string) string {
	// Generate context-aware NPC response
	return "Старейшина кивает вам и говорит: 'Добро пожаловать в наш город.'"
}

func (cg *CityGovernor) getCityName(cityID string) string {
	// Map city IDs to names
	switch cityID {
	case "city-ashes":
		return "Город Пепла"
	case "city-archives":
		return "Город Архивов"
	default:
		return "Неизвестный город"
	}
}

func (cg *CityGovernor) determineNextQuestType(playerID, cityID string) string {
	// Determine quest type based on player and city state
	return "help_citizen" // Simplified
}

func (cg *CityGovernor) generateQuestTitle(questType, cityID string) string {
	return "Помощь горожанину"
}

func (cg *CityGovernor) generateQuestDescription(questType, cityID string) string {
	return "Один из горожан просит вашей помощи с доставкой посылки."
}
