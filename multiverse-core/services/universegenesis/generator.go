// services/universegenesis/generator.go
package universegenesis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"multiverse-core/internal/eventbus"
	"multiverse-core/internal/oracle"
)

type Generator struct {
	bus       *eventbus.EventBus
	archivist *ArchivistClient // Теперь используется для сохранения universeBanProfile
	oracle    *oracle.Client   // <-- Изменён тип
}

func NewGenerator(bus *eventbus.EventBus, archivist *ArchivistClient, oracle *oracle.Client) *Generator {
	return &Generator{
		bus:       bus,
		archivist: archivist,
		oracle:    oracle,
	}
}

func (g *Generator) StartGenesis(ctx context.Context, seed string, constraints []string) error {
	log.Printf("Generating universe core laws and fundamental principles for seed: %s", seed)

	// 1. Генерация изначальных законов Вселенной и Ядра Вселенной через ИИ
	coreLaws, universeCore, err := g.generateUniverseCore(ctx, seed, constraints)
	if err != nil {
		return fmt.Errorf("failed to generate universe core: %w", err)
	}

	// 2. Генерация онтологического профиля для Запрета Вселенной через ИИ
	// Используем coreLaws и universeCore как контекст
	universeBanProfile, err := g.generateUniverseBanProfile(ctx, universeCore, coreLaws)
	if err != nil {
		return fmt.Errorf("failed to generate universe ban profile: %w", err)
	}

	// 3. Сохранение профиля онтологии для Запрета Вселенной
	profileJSON, err := json.Marshal(universeBanProfile)
	if err != nil {
		return fmt.Errorf("failed to marshal universe ban profile: %w", err)
	}

	// Сохраняем профиль с типом "universe_ontology_profile" и именем "cosmic_law"
	// Это даст путь в OntologicalArchivist: schemas/universe_ontology_profile/cosmic_law/1.0
	if err := g.archivist.SaveSchema(ctx, "universe_ontology_profile", "cosmic_law", "1.0", profileJSON); err != nil {
		log.Printf("Warning: Failed to save universe ban profile: %v", err)
		// Опять же, не критично для завершения генезиса, но желательно сохранить
	}
	log.Printf("Universe ban profile saved")

	core, err := json.Marshal(map[string]interface{}{
		"genesis_seed":  seed,
		"universe_core": universeCore,
		"cosmic_laws":   coreLaws,
		// Убираем archetypal_templates
	})

	// Это даст путь в OntologicalArchivist: schemas/universe_ontology_profile/cosmic_law/1.0
	if err := g.archivist.SaveSchema(ctx, "universe_core", "universe_core", "1.0", core); err != nil {
		log.Printf("Warning: Failed to save universe core: %v", err)
		// Опять же, не критично для завершения генезиса, но желательно сохранить
	}

	// 4. Генерация базовых схем для сущностей вселенной
	entityTypes := []string{"player", "npc", "house", "animal", "artifact"}
	for _, entityType := range entityTypes {
		// Используем функцию из worldgenerator для генерации схемы
		if err := GenerateEntitySchemaWithArchivist(g.archivist, ctx, entityType, seed); err != nil {
			log.Printf("Schema generation warning for %s: %v", entityType, err)
			// Continue with other types
		}
	}

	// Удаляем ненужную функцию

	// 5. Публикация финального события о завершении генезиса

	// Удаляем ненужную функцию

	// 5. Публикация финального события о завершении генезиса
	// Содержит только Ядро и Законы Вселенной
	finalEvent := eventbus.NewEvent(
		"universe.genesis.completed",
		"universe-genesis-oracle",
		seed,
		map[string]interface{}{
			"genesis_seed":  seed,
			"universe_core": universeCore,
			"cosmic_laws":   coreLaws,
			// Убираем archetypal_templates
		})
	/* 	finalEvent := eventbus.Event{
		EventID:   "universe-genesis-" + uuid.New().String()[:8],
		EventType: "universe.genesis.completed",
		Source:    "universe-genesis-oracle",
		WorldID:   "universe", // Условный ID для вселенной
		Payload: map[string]interface{}{
			"genesis_seed":  seed,
			"universe_core": universeCore,
			"cosmic_laws":   coreLaws,
			// Убираем archetypal_templates
		},
		Timestamp: time.Now(),
	} */
	g.bus.Publish(ctx, eventbus.TopicSystemEvents, finalEvent)

	log.Printf("Universe Genesis for seed '%s' completed successfully", seed)
	return nil
}

// generateUniverseCore вызывает Oracle для генерации изначальных законов и Ядра Вселенной.
func (g *Generator) generateUniverseCore(ctx context.Context, seed string, constraints []string) ([]string, string, error) {
	constraintsStr := ""
	if len(constraints) > 0 {
		constraintsStr = fmt.Sprintf("### ОГРАНИЧЕНИЯ: %s\n", constraints)
	}
	prompt := fmt.Sprintf(`Ты — Архитектор Вселенной. Создай изначальные законы и Ядро Вселенной на основе семени: "%s".
%s
Требуется сгенерировать:
1. Список изначальных законов/принципов (например, "Хаос как основа", "Ничто как потенциал", "Симметрия между материей и антиматерией" может быть что угодно инь янь, материя, мысль и т.д.).
2. Описание Ядра Вселенной (1-10 предложения, что лежит в основе всего).
Верни в JSON:
{
  "cosmic_laws": ["string"],
  "universe_core": "string"
} /no_think
`, seed, constraintsStr)

	// Создаём переменную нужного типа
	var result struct {
		CosmicLaws   []string `json:"cosmic_laws"`
		UniverseCore string   `json:"universe_core"`
	}

	// Вызываем метод, передавая указатель на переменную
	err := g.oracle.CallAndUnmarshal(ctx, func() (string, error) {
		return g.oracle.Call(ctx, prompt)
	}, &result)
	if err != nil {
		return nil, "", fmt.Errorf("oracle call for universe core failed: %w", err)
	}

	return result.CosmicLaws, result.UniverseCore, nil
}

// generateUniverseBanProfile вызывает Oracle для генерации онтологического профиля Запрета Вселенной.
func (g *Generator) generateUniverseBanProfile(ctx context.Context, universeCore string, cosmicLaws []string) (*OntologyProfile, error) {
	// Преобразуем []string в JSON строку для подстановки в промпт
	cosmicLawsJSON, err := json.Marshal(cosmicLaws)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cosmic laws for prompt: %w", err)
	}

	prompt := fmt.Sprintf(`Ты — Архитектор Онтологии Запрета Вселенной. На основе Ядра Вселенной и изначальных законов, создай онтологический профиль для Запрета Вселенной.
Это не строгая схема, а список возможных архетипов и принципов, которые Запрет Вселенной будет использовать для охраны изначальных законов.
ЯДРО ВСЕЛЕННОЙ: %s
ИЗНАЧАЛЬНЫЕ ЗАКОНЫ: %s
Определи:
- Возможные архетипы носителей (например, "Космический Закон", "Фундаментальная Сила", "Универсальный Принцип").
- Возможные архетипы сил (например, "Хаос", "Ничто", "Симметрия", "Порядок").
- Возможные архетипы связей (например, "Космический Резонанс", "Универсальная Нить").
- Возможные архетипы запретов (например, "Нарушение Симметрии", "Уничтожение Ничто", "Создание Абсолютного Порядка").
- Общие принципы (например, "Всё течёт из Источника и возвращается к нему", "Универсальные законы превалируют над локальными").
Верни это в виде JSON-описания профиля, не как строгую схему, а как список возможностей и ограничений.
{
  "archetypal_carriers": ["string"],
  "archetypal_forces": ["string"],
  "archetypal_connections": ["string"],
  "archetypal_forbiddances": ["string"],
  "general_principles": ["string"]
} /no_think
`, universeCore, string(cosmicLawsJSON))

	// Создаём переменную нужного типа
	var profile OntologyProfile

	// Вызываем метод, передавая указатель на переменную
	err = g.oracle.CallAndUnmarshal(ctx, func() (string, error) {
		return g.oracle.Call(ctx, prompt)
	}, &profile)
	if err != nil {
		return nil, fmt.Errorf("oracle call for universe ban profile failed: %w", err)
	}

	return &profile, nil
}

// --- Вспомогательные структуры ---

// OntologyProfile определяет структуру онтологического профиля
type OntologyProfile struct {
	ArchetypalCarriers     []string `json:"archetypal_carriers"`
	ArchetypalForces       []string `json:"archetypal_forces"`
	ArchetypalConnections  []string `json:"archetypal_connections"`
	ArchetypalForbiddances []string `json:"archetypal_forbiddances"`
	GeneralPrinciples      []string `json:"general_principles"`
}

// BaseEntitySchema is the base schema for all entities.
const BaseEntitySchema = `{
	"type": "object",
	"properties": {
		"entity_id": {"type": "string", "format": "entity_id"},
		"entity_type": {"type": "string"},
		"created_at": {"type": "string", "format": "date-time"},
		"updated_at": {"type": "string", "format": "date-time"},
		"payload": {"type": "object"},
		"history": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"event_id": {"type": "string", "format": "event_id"},
					"timestamp": {"type": "string", "format": "date-time"}
				},
				"required": ["event_id", "timestamp"],
				"additionalProperties": false
			}
		}
	},
	"required": ["entity_id", "entity_type", "created_at", "updated_at", "payload", "history"]
}`

// GenerateEntitySchemaWithArchivist generates and saves a schema for an entity type using provided archivist client
func GenerateEntitySchemaWithArchivist(archivist *ArchivistClient, ctx context.Context, entityType, worldSeed string) error {
	log.Printf("Generating schema for entity type: %s", entityType)

	// Generate payload schema via Oracle
	payloadSchemaStr, err := generatePayloadSchema(ctx, entityType, worldSeed)
	if err != nil {
		return fmt.Errorf("payload schema generation failed: %w", err)
	}

	// Parse base schema
	var baseSchema map[string]interface{}
	if err := json.Unmarshal([]byte(BaseEntitySchema), &baseSchema); err != nil {
		return fmt.Errorf("base schema parse failed: %w", err)
	}

	// Parse payload schema
	var payloadSchema map[string]interface{}
	if err := json.Unmarshal([]byte(payloadSchemaStr), &payloadSchema); err != nil {
		return fmt.Errorf("payload schema parse failed: %w", err)
	}

	// Merge schemas
	properties := baseSchema["properties"].(map[string]interface{})
	properties["payload"] = payloadSchema

	// Convert to bytes
	fullSchemaBytes, err := json.Marshal(baseSchema)
	if err != nil {
		return fmt.Errorf("schema marshal failed: %w", err)
	}

	// Save to OntologicalArchivist
	if err := archivist.SaveSchema(ctx, "entity", entityType, "1.0", fullSchemaBytes); err != nil {
		return fmt.Errorf("failed to save schema to archivist: %w", err)
	}

	log.Printf("Schema for %s saved to Archivist", entityType)
	return nil
}

// generatePayloadSchema asks Ascension Oracle to generate a payload schema.
func generatePayloadSchema(ctx context.Context, entityType, worldSeed string) (string, error) {
	prompt := fmt.Sprintf(`
 Сгенерируй ТОЛЬКО JSON Schema Draft 7 для поля "payload" сущности типа "%s" в мире с семенем "%s".

 Требования:
 1. Строго в формате JSON Schema.
 2. Используй "format": "entity_id" для ссылок на другие сущности.
 3. Укажи "required" поля.
 4. Пример для игрока:
 {
   "type": "object",
   "properties": {
     "name": {"type": "string"},
     "hp": {"type": "integer", "minimum": 0},
     "inventory": {
       "type": "array",
       "items": {"type": "string", "format": "entity_id"}
     }
   },
   "required": ["name", "hp"]
 }

 Верни ТОЛЬКО JSON без пояснений.
 `, entityType, worldSeed)

	client := oracle.NewClient()

	content, err := client.CallAndLog(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to connect to oracle: %w", err)
	}

	if content == "" {
		return "", fmt.Errorf("oracle returned empty content")
	}

	return content, nil
}
