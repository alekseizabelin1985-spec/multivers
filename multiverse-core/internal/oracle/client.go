// internal/oracle/client.go

package oracle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Client отвечает за взаимодействие с Ascension Oracle (Qwen3 и совместимые).
type Client struct {
	BaseURL string
	Model   string
	Client  *http.Client
	API_KEY string
}

// NewClient создаёт новый экземпляр клиента Oracle.
func NewClient() *Client {
	baseURL := os.Getenv("ORACLE_URL")
	if baseURL == "" {
		baseURL = "http://qwen3-service:11434/v1/chat/completions"
	}

	model := os.Getenv("ORACLE_MODEL")
	if model == "" {
		model = "qwen3"
		log.Printf("ORACLE_MODEL not set, using default: %s", model)
	}

	apiKey := os.Getenv("ORACLE_API_KEY")
	if apiKey == "" {
		log.Println("WARNING: ORACLE_API_KEY not set — requests may fail")
	}

	timeoutMsStr := os.Getenv("ORACLE_TIMEOUT_MS")
	timeoutMs := 10000 // default: 10 seconds
	if t, err := strconv.Atoi(timeoutMsStr); err == nil && t > 0 {
		timeoutMs = t
	}
	timeout := time.Duration(timeoutMs) * time.Millisecond

	return &Client{
		BaseURL: baseURL,
		Model:   model,
		Client: &http.Client{
			Timeout: timeout,
		},
		API_KEY: apiKey,
	}
}

// doRequest выполняет HTTP-запрос с retry для временных ошибок.
func (c *Client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt < 3; attempt++ {
		resp, err = c.Client.Do(req.WithContext(ctx))
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		if resp != nil {
			resp.Body.Close()
		}

		if attempt < 2 {
			sleepMs := 100 * (1 << attempt) // 100ms, 200ms
			time.Sleep(time.Duration(sleepMs) * time.Millisecond)
		}
	}
	return resp, err
}

// callRaw выполняет вызов Oracle с произвольным телом запроса.
func (c *Client) callRaw(ctx context.Context, requestBody []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.API_KEY != "" {
		req.Header.Set("Authorization", "Bearer "+c.API_KEY)
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Проверка Content-Type
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(strings.ToLower(ct), "application/json") {
		return "", fmt.Errorf("unexpected content type: %s", ct)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("oracle returned status %d: %s", resp.StatusCode, string(body))
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Безопасное логирование
	log.Printf("Oracle response: %d bytes, status %d", len(responseBody), resp.StatusCode)

	var responseStruct struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(responseBody, &responseStruct); err != nil {
		return "", fmt.Errorf("failed to unmarshal oracle response: %w", err)
	}

	if len(responseStruct.Choices) == 0 {
		return "", fmt.Errorf("oracle returned no choices")
	}
	content := responseStruct.Choices[0].Message.Content
	if content == "" {
		return "", fmt.Errorf("oracle returned empty content")
	}

	return content, nil
}

// CallStructured вызывает Oracle с system и user промтами (рекомендуемый метод).
// systemPrompt — контекст (факты, сущности, окружение),
// userPrompt — задача и требования.
func (c *Client) CallStructured(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"model": c.Model,
		"messages": []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.7,
		"max_tokens":  1024,
		"response_format": map[string]string{
			"type": "json_object",
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal prompt: %w", err)
	}

	return c.callRaw(ctx, requestBody)
}

// Call — устаревший метод для обратной совместимости.
// Передаёт весь промт как один user-месседж.
// Рекомендуется использовать CallStructured.
func (c *Client) Call(ctx context.Context, prompt string) (string, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"model": c.Model,
		"messages": []map[string]interface{}{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.8,
		"max_tokens":  1500,
		"response_format": map[string]string{
			"type": "json_object",
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal prompt: %w", err)
	}

	return c.callRaw(ctx, requestBody)
}

// CallAndUnmarshal вызывает Oracle и десериализует ответ в target.
// Поддерживает оба метода — просто передайте нужную функцию.
// Пример:
//
//	var result NarrativeResponse
//	err := c.CallAndUnmarshal(ctx, func() (string, error) {
//	    return c.CallStructured(ctx, system, user)
//	}, &result)
func (c *Client) CallAndUnmarshal(ctx context.Context, callFunc func() (string, error), target interface{}) error {
	response, err := callFunc()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(response), target)
}

// CallAndLog вызывает Oracle и логирует ошибки (устаревший интерфейс).
func (c *Client) CallAndLog(ctx context.Context, prompt string) (string, error) {
	return c.Call(ctx, prompt)
}
