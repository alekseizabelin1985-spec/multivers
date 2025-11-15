// Package worldgenerator implements world generation logic.
package worldgenerator

import (
	"context"
	"fmt"
	"log"
	"time"

	"multiverse-core/internal/eventbus"

	"github.com/google/uuid"
)

// WorldGeography представляет полную географическую структуру мира
type WorldGeography struct {
	Core      string        `json:"core"`
	Ontology  WorldOntology `json:"ontology"`
	Geography Geography     `json:"geography"`
	Mythology string        `json:"mythology"`
}

// WorldOntology представляет онтологию культивации мира
type WorldOntology struct {
	Carriers  []string `json:"carriers"`
	Paths     []string `json:"paths"`
	Forbidden []string `json:"forbidden"`
}

// Geography представляет географическую структуру мира
type Geography struct {
	Regions     []Region    `json:"regions"`
	WaterBodies []WaterBody `json:"water_bodies"`
	Cities      []City      `json:"cities"`
}

// Region представляет регион мира
type Region struct {
	Name        string  `json:"name"`
	Biome       string  `json:"biome"`
	Coordinates Point   `json:"coordinates"`
	Size        float64 `json:"size"`
}

// WaterBody представляет водный объект
type WaterBody struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"` // river, sea, lake
	Coordinates Point   `json:"coordinates"`
	Size        float64 `json:"size"`
}

// City представляет город
type City struct {
	Name       string   `json:"name"`
	Population int      `json:"population"`
	Type       string   `json:"type"` // major, minor
	Location   Location `json:"location"`
}

// Location представляет местоположение
type Location struct {
	Region      string `json:"region"`
	Coordinates Point  `json:"coordinates"`
}

// Point представляет точку координат
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// WorldGenerator creates new worlds with unique ontologies.
type WorldGenerator struct {
	bus       *eventbus.EventBus
	archivist ArchivistClient
}

// NewWorldGenerator creates a new WorldGenerator.
func NewWorldGenerator(bus *eventbus.EventBus) *WorldGenerator {
	return &WorldGenerator{
		bus:       bus,
		archivist: *NewArchivistClient(),
	}
}

// HandleEvent processes world generation requests.
func (wg *WorldGenerator) HandleEvent(ev eventbus.Event) {
	if ev.EventType != "world.generation.requested" {
		return
	}

	worldSeed, _ := ev.Payload["seed"].(string)
	if worldSeed == "" {
		log.Printf("Invalid world generation request: missing seed")
		return
	}

	log.Printf("Starting world generation: %s", worldSeed)


	// Create world entity
	worldID := "world-" + uuid.New().String()[:8]
	worldEvent := eventbus.Event{
		EventID:   "world-gen-" + uuid.New().String()[:8],
		EventType: "entity.created",
		Source:    "world-generator",
		WorldID:   worldID,
		Payload: map[string]interface{}{
			"entity_id":   worldID,
			"entity_type": "world",
			"payload": map[string]interface{}{
				"seed":        worldSeed,
				"plan":        0,
				"core":        "",
				"constraints": ev.Payload["constraints"],
			},
		},
		Timestamp: time.Now(),
	}
	wg.bus.Publish(context.Background(), eventbus.TopicSystemEvents, worldEvent)

	// Generate world details via Oracle
	wg.generateEnhancedWorldDetails(context.Background(), worldID, worldSeed)

	// Publish world generated event
	finalEvent := eventbus.Event{
		EventID:   "world-gen-final-" + uuid.New().String()[:8],
		EventType: "world.generated",
		Source:    "world-generator",
		WorldID:   worldID,
		Payload: map[string]interface{}{
			"world_id": worldID,
			"seed":     worldSeed,
		},
		Timestamp: time.Now(),
	}
	wg.bus.Publish(context.Background(), eventbus.TopicSystemEvents, finalEvent)

	log.Printf("World %s generated successfully", worldID)
}

// generateEnhancedWorldDetails generates enhanced world details via Ascension Oracle.
func (wg *WorldGenerator) generateEnhancedWorldDetails(ctx context.Context, worldID, worldSeed string) {
	prompt := fmt.Sprintf(`
Создай детали мира с семенем "%s".

Требуется сгенерировать:
1. Ядро Мира (1-2 предложения)
2. Онтологию культивации (носители, пути, запреты)
3. Географию:
   - 3-5 регионов с уникальными биомами (леса, горы, поля, пустыни, болота)
   - Водные объекты (2-4 реки, 1-2 моря, 1-3 озера)
   - 2-4 города с основными характеристиками
4. Мифологию (краткий миф)

Верни строго в JSON:
{
  "core": "string",
  "ontology": {
    "carriers": ["string"],
    "paths": ["string"],
    "forbidden": ["string"]
  },
  "geography": {
    "regions": [
      {
        "name": "string",
        "biome": "string",
        "coordinates": {"x": 0.0, "y": 0.0},
        "size": 0.0
      }
    ],
    "water_bodies": [
      {
        "name": "string",
        "type": "river|sea|lake",
        "coordinates": {"x": 0.0, "y": 0.0},
        "size": 0.0
      }
    ],
    "cities": [
      {
        "name": "string",
        "population": 0,
        "type": "major|minor",
        "location": {
          "region": "string",
          "coordinates": {"x": 0.0, "y": 0.0}
        }
      }
    ]
  },
  "mythology": "string"
}
`, worldSeed)

	// resp, err := CallOracle(ctx, prompt)
	// if err != nil {
	// 	log.Printf("Oracle world details failed: %v", err)
	// 	return
	// }

	// Parse the response
	var geography WorldGeography

	err := CallOracleAndUnmarshal(ctx, prompt, &geography)

	if err != nil {
		log.Printf("Oracle world details failed: %v", err)
		return
	}

	// Oracle returns JSON with "narrative" field containing the actual JSON data
	// First, we need to unmarshal the narrative field to get the actual JSON
	// Clean up potential markdown formatting

	// Create geographic entities and publish events
	wg.createGeographicEntities(ctx, worldID, geography)
}

// createGeographicEntities creates entities for geographic objects
func (wg *WorldGenerator) createGeographicEntities(ctx context.Context, worldID string, geography WorldGeography) {
	// Create regions
	for _, region := range geography.Geography.Regions {
		wg.createRegionEntity(ctx, worldID, region)
	}

	// Create water bodies
	for _, water := range geography.Geography.WaterBodies {
		wg.createWaterEntity(ctx, worldID, water)
	}

	// Create cities
	for _, city := range geography.Geography.Cities {
		wg.createCityEntity(ctx, worldID, city)
	}

	// Publish geography generated event
	wg.publishGeographyGeneratedEvent(ctx, worldID, geography)
}

// createRegionEntity creates a region entity
func (wg *WorldGenerator) createRegionEntity(ctx context.Context, worldID string, region Region) {
	regionID := "region-" + uuid.New().String()[:8]

	regionEvent := eventbus.Event{
		EventID:   "region-create-" + uuid.New().String()[:8],
		EventType: "entity.created",
		Source:    "world-generator",
		WorldID:   worldID,
		Payload: map[string]interface{}{
			"entity_id":   regionID,
			"entity_type": "region",
			"payload": map[string]interface{}{
				"name":        region.Name,
				"biome":       region.Biome,
				"coordinates": region.Coordinates,
				"size":        region.Size,
			},
		},
		Timestamp: time.Now(),
	}

	wg.bus.Publish(ctx, eventbus.TopicSystemEvents, regionEvent)
	log.Printf("Created region entity: %s (%s)", region.Name, region.Biome)
}

// createWaterEntity creates a water body entity
func (wg *WorldGenerator) createWaterEntity(ctx context.Context, worldID string, water WaterBody) {
	waterID := "water-" + uuid.New().String()[:8]

	waterEvent := eventbus.Event{
		EventID:   "water-create-" + uuid.New().String()[:8],
		EventType: "entity.created",
		Source:    "world-generator",
		WorldID:   worldID,
		Payload: map[string]interface{}{
			"entity_id":   waterID,
			"entity_type": "water_body",
			"payload": map[string]interface{}{
				"name":        water.Name,
				"type":        water.Type,
				"coordinates": water.Coordinates,
				"size":        water.Size,
			},
		},
		Timestamp: time.Now(),
	}

	wg.bus.Publish(ctx, eventbus.TopicSystemEvents, waterEvent)
	log.Printf("Created water entity: %s (%s)", water.Name, water.Type)
}

// createCityEntity creates a city entity
func (wg *WorldGenerator) createCityEntity(ctx context.Context, worldID string, city City) {
	cityID := "city-" + uuid.New().String()[:8]

	cityEvent := eventbus.Event{
		EventID:   "city-create-" + uuid.New().String()[:8],
		EventType: "entity.created",
		Source:    "world-generator",
		WorldID:   worldID,
		Payload: map[string]interface{}{
			"entity_id":   cityID,
			"entity_type": "city",
			"payload": map[string]interface{}{
				"name":       city.Name,
				"population": city.Population,
				"type":       city.Type,
				"location":   city.Location,
			},
		},
		Timestamp: time.Now(),
	}

	wg.bus.Publish(ctx, eventbus.TopicSystemEvents, cityEvent)
	log.Printf("Created city entity: %s (population: %d)", city.Name, city.Population)
}

// publishGeographyGeneratedEvent publishes an event when geography is generated
func (wg *WorldGenerator) publishGeographyGeneratedEvent(ctx context.Context, worldID string, geography WorldGeography) {
	geographyEvent := eventbus.Event{
		EventID:   "geography-generated-" + uuid.New().String()[:8],
		EventType: "world.geography.generated",
		Source:    "world-generator",
		WorldID:   worldID,
		Payload: map[string]interface{}{
			"world_id":     worldID,
			"regions":      len(geography.Geography.Regions),
			"water_bodies": len(geography.Geography.WaterBodies),
			"cities":       len(geography.Geography.Cities),
		},
		Timestamp: time.Now(),
	}

	wg.bus.Publish(ctx, eventbus.TopicSystemEvents, geographyEvent)
	log.Printf("Published geography generated event for world: %s", worldID)
}
