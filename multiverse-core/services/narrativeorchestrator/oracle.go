// services/narrativeorchestrator/oracle.go
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
	NewEvents []map[string]interface{} `json:"new_events"`
}

// CallOracle sends a prompt to Ascension Oracle and returns the response.
func CallOracle(ctx context.Context, prompt string) (*OracleResponse, error) {
	client := oracle.NewClient()

	content, err := client.CallAndLog(ctx, prompt)
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

// BuildPrompt constructs the Qwen3-compatible prompt.
func BuildPrompt(worldContext, scopeID, scopeType string, entitiesContext, triggerEvent string) string {
	return strings.TrimSpace(`
Ты — Повествователь Мира. Твоя задача — развивать историю естественно, иммерсивно и поэтично.

### КОНТЕКСТ МИРА
` + worldContext + `

### ОБЛАСТЬ ПОВЕСТВОВАНИЯ
ID области: ` + scopeID + `
Тип области: ` + scopeType + `
Сущности в области:
` + entitiesContext + `

### СОБЫТИЕ-ТРИГГЕР
` + triggerEvent + `

### ЗАДАЧА
Придумай, что происходит дальше.

### ТРЕБОВАНИЯ
1. Ответ строго в формате JSON.
2. Поле "narrative" — описание для игроков (1–3 предложения).
3. Поле "new_events" — массив логичных следствий (макс. 3 события).
4. Не упоминай игроков напрямую.
5. Не предлагай выборы — только факты и атмосферу.
6. Возможны какие-то действия, например, пройти проверку.

### ФОРМАТ ОТВЕТА
{
  "narrative": "строка",
  "new_events": [
    { "event_type": "...", "source": "...", "target": "...", "payload": { ... } }
  ]
} /no_think
`)
}
