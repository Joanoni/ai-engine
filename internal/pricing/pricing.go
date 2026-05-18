package pricing

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ModelPrice holds the per-million-token prices for a model.
type ModelPrice struct {
	InputPerMillion  float64 `json:"input_per_million"`
	OutputPerMillion float64 `json:"output_per_million"`
	Currency         string  `json:"currency"`
}

// Load reads .ai-engine/model-pricing.json from the workspace.
// Returns an empty map (not an error) if the file does not exist.
func Load(workspacePath string) (map[string]ModelPrice, error) {
	path := filepath.Join(workspacePath, ".ai-engine", "model-pricing.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]ModelPrice{}, nil
		}
		return nil, err
	}
	var prices map[string]ModelPrice
	if err := json.Unmarshal(data, &prices); err != nil {
		return nil, err
	}
	return prices, nil
}

// CalcCost returns the estimated cost in USD for the given token counts.
// Returns 0 if the model is not in the pricing map.
func CalcCost(prices map[string]ModelPrice, model string, inputTokens, outputTokens int) float64 {
	p, ok := prices[model]
	if !ok {
		return 0
	}
	inputCost := float64(inputTokens) / 1_000_000 * p.InputPerMillion
	outputCost := float64(outputTokens) / 1_000_000 * p.OutputPerMillion
	return inputCost + outputCost
}
