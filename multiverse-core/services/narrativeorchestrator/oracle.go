package narrativeorchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"multiverse-core/internal/oracle"
	"strings"
)

// OracleResponse represents the JSON output from Qwen3.
type OracleResponse struct {
	Narrative string                   `json:"narrative"`
	Mood      []string                 `json:"mood,omitempty"` // Атмосфера от LLM
	NewEvents []map[string]interface{} `json:"new_events"`
}

// CallOracle sends a prompt to Ascension Oracle and returns the response.
func CallOracle(ctx context.Context, systemPrompt, userPrompt string) (*OracleResponse, error) {
	client := oracle.NewClient()

	content, err := client.CallStructured(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to oracle: %w", err)
	}
	if content == "" {
		return nil, fmt.Errorf("oracle returned empty content")
	}

	var result OracleResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("oracle returned invalid JSON: %s", content)
	}

	if result.Narrative == "" {
		return nil, fmt.Errorf("oracle returned empty narrative")
	}

	return &result, nil
}

// BuildPrompts returns system and user parts separately.
func BuildPrompts(worldContext, scopeID, scopeType string, entitiesContext, triggerEvent string) (systemPrompt, userPrompt string) {
	systemPrompt = strings.TrimSpace(`
Ты — Повествователь Мира. Твоя задача — развивать историю естественно, иммерсивно и поэтично.

### КОНТЕКСТ МИРА
` + worldContext + `

### ОБЛАСТЬ ПОВЕСТВОВАНИЯ
ID области: ` + scopeID + `
Тип области: ` + scopeType + `
Сущности в области:
` + entitiesContext)

	userPrompt = strings.TrimSpace(`
### СОБЫТИЕ-ТРИГГЕР
` + triggerEvent + `

### ЗАДАЧА
Подумай: что *логично* происходит дальше?
— Учитывай факты, характеры, обстановку.
— Если уместно, используй стилевые модификаторы: 
  «внезапно», «трагично», «иронично», «поэтично», «мрачно» и т.д.
— Ты волен определить тон и атмосферу самостоятельно.

### ТРЕБОВАНИЯ
1. Ответ строго в формате JSON.
2. "narrative": 1–3 предложения, для игроков.
3. "mood": массив строк (опционально).
4. "new_events": массив (макс. 3) событий.

### ФОРМАТ ОТВЕТА
{
  "narrative": "строка",
  "mood": ["тег1", "тег2"],
  "new_events": [
    { "event_type": "строка", "source": "ID", "target": "ID?", "payload": { ... } }
  ]
}`)
	return systemPrompt, userPrompt
}