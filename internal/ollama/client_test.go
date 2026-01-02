package ollama

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:11434")

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.baseURL != "http://localhost:11434" {
		t.Errorf("NewClient() baseURL = %q, want %q", client.baseURL, "http://localhost:11434")
	}
}

func TestNewClientDefault(t *testing.T) {
	client := NewClientDefault()

	if client == nil {
		t.Fatal("NewClientDefault() returned nil")
	}

	if client.baseURL != DefaultBaseURL {
		t.Errorf("NewClientDefault() baseURL = %q, want %q", client.baseURL, DefaultBaseURL)
	}
}

func TestClient_IsHealthy(t *testing.T) {
	// Create mock server that responds to health check
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Ollama is running"))
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthy := client.IsHealthy(ctx)
	if !healthy {
		t.Error("IsHealthy() = false, want true for running server")
	}
}

func TestClient_IsHealthy_Unhealthy(t *testing.T) {
	// Use invalid URL to simulate unhealthy
	client := NewClient("http://localhost:99999")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	healthy := client.IsHealthy(ctx)
	if healthy {
		t.Error("IsHealthy() = true, want false for non-existent server")
	}
}

func TestClient_ListModels(t *testing.T) {
	// Create mock server that returns model list
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"models": [
					{"name": "llama3:latest", "size": 4000000000},
					{"name": "mistral:latest", "size": 3000000000}
				]
			}`))
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	models, err := client.ListModels(ctx)
	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}

	if len(models) != 2 {
		t.Errorf("ListModels() returned %d models, want 2", len(models))
	}

	if models[0].Name != "llama3:latest" {
		t.Errorf("ListModels()[0].Name = %q, want %q", models[0].Name, "llama3:latest")
	}
}

func TestClient_ListModels_Error(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.ListModels(ctx)
	if err == nil {
		t.Error("ListModels() should return error for 500 response")
	}
}

func TestModel_String(t *testing.T) {
	model := Model{
		Name: "llama3:latest",
		Size: 4_000_000_000,
	}

	str := model.String()
	if !strings.Contains(str, "llama3") {
		t.Errorf("Model.String() = %q, want to contain 'llama3'", str)
	}
}
