// services/narrativeorchestrator/oracle.go

package narrativeorchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"multiverse-core/internal/oracle"
	"strings"
	"time"
)

type OracleResponse struct {
	Narrative string                   `json:"narrative"`
	Mood      []string                 `json:"mood,omitempty"`
	NewEvents []map[string]interface{} `json:"new_events"`
}

// PromptInput — данные для генерации промта.
type PromptInput struct {
	WorldContext    string
	ScopeID         string
	ScopeType       string
	EntitiesContext string
	EventClusters   []EventCluster
	TimeContext     string
	TriggerEvent    string
}

// BuildPrompt генерирует промт для Oracle.
func BuildPrompt(input PromptInput) (string, string) {
	system := strings.TrimSpace(`
Ты — Повествователь Мира. Твоя задача — развивать историю естественно, иммерсивно и поэтично.

### КОНТЕКСТ МИРА
` + input.WorldContext + `

### ОБЛАСТЬ ПОВЕСТВОВАНИЯ
ID области: ` + input.ScopeID + `
Тип области: ` + input.ScopeType + `
Сущности в области:
` + input.EntitiesContext)

	user := strings.TrimSpace(`
### ВРЕМЕННОЙ КОНТЕКСТ
` + input.TimeContext + `

### НАКОПЛЕННЫЕ СОБЫТИЯ
` + buildEventClusters(input.EventClusters) + `

### СОБЫТИЕ-ТРИГГЕР
` + input.TriggerEvent + `

### ЗАДАЧА
Подумай: что *логично* происходит дальше?
— Учитывай факты, характеры, обстановку.
— Даже если событий мало — мир живёт.
— Используй стилевые модификаторы: «внезапно», «плавно», «тревожно».

### СОЗДАНИЕ И ОБНОВЛЕНИЕ СУЩНОСТЕЙ
Генерируй события ТОЛЬКО в формате EntityManager.

### ТРЕБОВАНИЯ
1. Ответ строго в формате JSON.
2. "narrative": 1–3 предложения.
3. "mood": массив строк (опционально).
4. "new_events": массив (макс. 3) событий.`)
	return system, user
}

func CallOracle(ctx context.Context, systemPrompt, userPrompt string) (*OracleResponse, error) {
	client := oracle.NewClient()
	content, err := client.CallStructured(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("oracle call failed: %w", err)
	}
	if content == "" {
		return nil, fmt.Errorf("empty content")
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

// Вспомогательные функции
func buildEventClusters(clusters []EventCluster) string {
	if len(clusters) == 0 {
		return "Нет событий за период.\n"
	}
	var lines []string
	for _, c := range clusters {
		lines = append(lines, fmt.Sprintf("[ %s ] %s", c.RelativeTime, c.Description))
	}
	return strings.Join(lines, "\n")
}

func BuildTimeContext(lastEventTime *time.Time, lastMood []string) string {
	var lines []string
	now := time.Now()
	lines = append(lines, "Абсолютное время: "+now.Format("15:04, 02.01.2006"))
	lines = append(lines, "- Сутки: "+getDayPhase(now))
	lines = append(lines, "- Сезон: "+getSeason(now))

	if lastEventTime != nil {
		ago := now.Sub(*lastEventTime)
		agoDesc := humanizeDuration(int64(ago.Milliseconds()))
		lines = append(lines, fmt.Sprintf("- Последнее событие: %s", agoDesc))
	}

	if len(lastMood) > 0 {
		lines = append(lines, fmt.Sprintf("- Атмосфера: %s", strings.Join(lastMood, ", ")))
	}

	return strings.Join(lines, "\n")
}

func humanizeDuration(dt int64) string {
	switch {
	case dt <= 0:
		return "одновременно"
	case dt <= 50:
		return "почти одновременно"
	case dt <= 200:
		return "мгновенно после"
	case dt <= 800:
		return "через мгновение"
	case dt <= 1500:
		return "через секунду"
	case dt <= 3000:
		return "спустя пару секунд"
	case dt <= 10000:
		return "спустя несколько секунд"
	default:
		secs := dt / 1000
		return fmt.Sprintf("спустя %d секунд", secs)
	}
}

func getDayPhase(t time.Time) string {
	h := t.Hour()
	switch {
	case h >= 5 && h < 12:
		return "утро"
	case h >= 12 && h < 18:
		return "день"
	case h >= 18 && h < 22:
		return "вечер"
	default:
		return "ночь"
	}
}

func getSeason(t time.Time) string {
	month := t.Month()
	switch month {
	case time.December, time.January, time.February:
		return "зима"
	case time.March, time.April, time.May:
		return "весна"
	case time.June, time.July, time.August:
		return "лето"
	default:
		return "осень"
	}
}
