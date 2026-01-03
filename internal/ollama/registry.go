package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// RegistryModel represents a model from the registry.
type RegistryModel struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Fallback list of popular models
var fallbackModels = []RegistryModel{
	{"llama3.2", "Meta's latest, 3B params"},
	{"llama3.2:1b", "Lightweight, 1B params"},
	{"llama3.1", "Meta Llama 3.1, 8B params"},
	{"mistral", "Mistral 7B, fast & capable"},
	{"gemma3", "Google Gemma 3"},
	{"phi4", "Microsoft Phi-4, 14B"},
	{"qwen3", "Alibaba Qwen 3"},
	{"deepseek-r1", "DeepSeek reasoning model"},
	{"codellama", "Code generation, 7B"},
	{"llava", "Vision + Language model"},
	{"nomic-embed-text", "Text embeddings"},
}

// FetchAvailableModels tries external API, falls back to hardcoded list.
func FetchAvailableModels(ctx context.Context) []RegistryModel {
	// Try external API with short timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	models, err := fetchFromAPI(ctxTimeout)
	if err == nil && len(models) > 0 {
		return models
	}

	// Fallback to hardcoded list
	return fallbackModels
}

func fetchFromAPI(ctx context.Context) ([]RegistryModel, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://ollamadb.dev/api/v1/models?limit=20", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Models []RegistryModel `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Models, nil
}
