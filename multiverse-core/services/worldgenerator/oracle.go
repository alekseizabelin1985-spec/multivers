// Package worldgenerator handles communication with Ascension Oracle.
package worldgenerator

import (
	"context"
	"fmt"
	"multiverse-core/internal/oracle"
)

// OracleResponse represents the JSON output from Qwen3.
type OracleResponse struct {
	Narrative string `json:"narrative"`
}

// CallOracle sends a prompt to Ascension Oracle and returns the response.
func CallOracle(ctx context.Context, prompt string) (string, error) {

	client := oracle.NewClient()

	content, err := client.CallAndLog(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to connect to oracle: %w", err)
	}

	if content == "" {
		return "", fmt.Errorf("oracle returned empty content")
	}

	return content, nil
	// // Clean up markdown code blocks
	// content = strings.Trim(content, "` \n")
	// if strings.HasPrefix(content, "json") {
	// 	content = strings.TrimSpace(content[4:])
	// }

	// var result OracleResponse
	// if err := json.Unmarshal([]byte(content), &result); err != nil {
	// 	log.Printf("Oracle returned invalid JSON, using as narrative: %s", content)
	// 	result.Narrative = content
	// }

	// return &result, nil
}

// CallOracle sends a prompt to Ascension Oracle and returns the response.
func CallOracleAndUnmarshal(ctx context.Context, prompt string, target interface{}) error {

	client := oracle.NewClient()

	err := client.CallAndUnmarshal(ctx, func() (string, error) {
	    return client.Call(ctx, prompt)
	}  , target)
	if err != nil {
		return fmt.Errorf("failed to connect to oracle: %w", err)
	}

	return nil
}
