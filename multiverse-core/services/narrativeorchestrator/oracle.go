// services/narrativeorchestrator/oracle.go

package narrativeorchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"multiverse-core/internal/oracle"
	"strings"
)

type OracleResponse struct {
	Narrative string                   `json:"narrative"`
	Mood      []string                 `json:"mood,omitempty"`
	NewEvents []map[string]interface{} `json:"new_events"`
}

func CallOracle(ctx context.Context, systemPrompt, userPrompt string) (*OracleResponse, error) {
	client := oracle.NewClient()
	content, err := client.CallStructured(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("oracle call failed: %w", err)
	}
	if content == "" {
		return nil, fmt.Errorf("oracle returned empty content")
	}

	var result OracleResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("invalid JSON: %s", content)
	}
	if result.Narrative == "" {
		return nil, fmt.Errorf("empty narrative")
	}
	return &result, nil
}

func BuildPrompts(worldContext, scopeID, scopeType string, entitiesContext, triggerEvent string) (string, string) {
	system := strings.TrimSpace(`
Ты — Повествователь Мира. Твоя задача — развивать историю естественно, иммерсивно и поэтично.

### КОНТЕКСТ МИРА
` + worldContext + `

### ОБЛАСТЬ ПОВЕСТВОВАНИЯ
ID области: ` + scopeID + `
Тип области: ` + scopeType + `
Сущности в области:
` + entitiesContext)

	user := strings.TrimSpace(`
### СОБЫТИЕ-ТРИГГЕР
` + triggerEvent + `

### ЗАДАЧА
Подумай: что *логично* происходит дальше?
— Учитывай факты, характеры, обстановку.
— Даже если событий мало — мир живёт: ветер, тени, эмоции.
— Используй стилевые модификаторы: «внезапно», «плавно», «тревожно».

### СОЗДАНИЕ И ОБНОВЛЕНИЕ СУЩНОСТЕЙ
Генерируй события ТОЛЬКО в формате EntityManager:

1. Для новых сущностей:
   { "event_type": "entity.created", "entity_id": "...", "entity_type": "...", "payload": { ... } }

2. Для обновления:
   {
     "event_type": "world_events",
     "payload": {
       "state_changes": [
         { "entity_id": "...", "operations": [ { "op": "set", "path": "...", "value": ... } ] }
       ]
     }
   }

### ТРЕБОВАНИЯ
1. Ответ строго в формате JSON.
2. "narrative": 1–3 предложения.
3. "mood": массив строк (опционально).
4. "new_events": массив (макс. 3) событий.

### ФОРМАТ ОТВЕТА
{
  "narrative": "Тень сгустилась — и из неё выскочил пёс!",
  "mood": ["sudden", "threatening"],
  "new_events": [
    {
      "event_type": "entity.created",
      "entity_id": "npc:shadow_hound-777",
      "entity_type": "npc",
      "payload": { "name": "Теневой пёс", "hp": 45 }
    },
    {
      "event_type": "world_events",
      "payload": {
        "state_changes": [{
          "entity_id": "player:kain-777",
          "operations": [
            { "op": "set", "path": "perception", "value": 1.5 }
          ]
        }]
      }
    }
  ]
}`)
	return system, user
}
