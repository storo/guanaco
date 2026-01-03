package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Message represents a chat message.
type Message struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

// ChatRequest represents a request to the chat API.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// chatResponse represents a streaming response chunk from the chat API.
type chatResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done  bool   `json:"done"`
	Error string `json:"error,omitempty"`
}

// TokenCallback is called for each token received during streaming.
type TokenCallback func(token string)

// StreamHandler handles streaming chat responses from Ollama.
type StreamHandler struct {
	client *Client
}

// NewStreamHandler creates a new stream handler.
func NewStreamHandler(client *Client) *StreamHandler {
	return &StreamHandler{
		client: client,
	}
}

// Chat sends a chat request and streams the response tokens.
// The callback is called for each token received.
// Returns when the response is complete or context is cancelled.
func (h *StreamHandler) Chat(ctx context.Context, req *ChatRequest, callback TokenCallback) error {
	// Always stream
	req.Stream = true

	// Encode request body
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	// Create HTTP request
	url := h.client.baseURL + "/api/chat"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Use a client without timeout for streaming (model loading can take time)
	streamClient := &http.Client{}
	resp, err := streamClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var chunk chatResponse
		if err := json.Unmarshal(line, &chunk); err != nil {
			// Skip malformed lines
			continue
		}

		// Check for error in response
		if chunk.Error != "" {
			return fmt.Errorf("ollama error: %s", chunk.Error)
		}

		// Call callback with token
		if chunk.Message.Content != "" {
			callback(chunk.Message.Content)
		}

		// Check if done
		if chunk.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		// Check if it was a context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return fmt.Errorf("error reading response: %w", err)
		}
	}

	return nil
}
