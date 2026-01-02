// Package ollama provides a client for the Ollama API.
package ollama

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default Ollama API endpoint.
	DefaultBaseURL = "http://localhost:11434"

	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second
)

// Model represents an Ollama model.
type Model struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
}

// String returns a human-readable representation of the model.
func (m Model) String() string {
	sizeMB := m.Size / (1024 * 1024)
	return fmt.Sprintf("%s (%d MB)", m.Name, sizeMB)
}

// modelsResponse is the API response for listing models.
type modelsResponse struct {
	Models []Model `json:"models"`
}

// Client is an HTTP client for the Ollama API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Ollama client with the given base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// NewClientDefault creates a new Ollama client with the default base URL.
func NewClientDefault() *Client {
	return NewClient(DefaultBaseURL)
}

// IsHealthy checks if the Ollama server is running and responsive.
func (c *Client) IsHealthy(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// ListModels returns all available models from the Ollama server.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	url := c.baseURL + "/api/tags"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var modelsResp modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return modelsResp.Models, nil
}

// PullProgressCallback is called with progress updates during model pull.
type PullProgressCallback func(status string, completed, total int64)

// PullModel downloads a model from the Ollama registry.
func (c *Client) PullModel(ctx context.Context, model string, callback PullProgressCallback) error {
	url := c.baseURL + "/api/pull"

	body := fmt.Sprintf(`{"name": "%s", "stream": true}`, model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Use a client without timeout for long downloads
	pullClient := &http.Client{}
	resp, err := pullClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read streaming progress
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var progress struct {
			Status    string `json:"status"`
			Completed int64  `json:"completed"`
			Total     int64  `json:"total"`
			Error     string `json:"error"`
		}

		if err := json.Unmarshal(line, &progress); err != nil {
			continue
		}

		if progress.Error != "" {
			return fmt.Errorf("pull error: %s", progress.Error)
		}

		if callback != nil {
			callback(progress.Status, progress.Completed, progress.Total)
		}
	}

	return scanner.Err()
}

// HasModel checks if a model is available locally.
func (c *Client) HasModel(ctx context.Context, model string) bool {
	models, err := c.ListModels(ctx)
	if err != nil {
		return false
	}

	for _, m := range models {
		if m.Name == model || strings.HasPrefix(m.Name, model+":") {
			return true
		}
	}
	return false
}
