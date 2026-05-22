package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

var baseURL string

func TestMain(m *testing.M) {
	baseURL = os.Getenv("CONTINUUM_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Wait for API to be ready
	if err := waitForAPI(); err != nil {
		fmt.Printf("API not ready: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func waitForAPI() error {
	client := &http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 30; i++ {
		resp, err := client.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("API did not become ready")
}

func TestHealthEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAgentLifecycle(t *testing.T) {
	// Create agent
	agent := map[string]interface{}{
		"name":       "test-agent",
		"model":      "gpt-4",
		"provider":   "openai",
		"max_tokens": 100000,
	}

	body, _ := json.Marshal(agent)
	resp, err := http.Post(baseURL+"/api/v1/agents", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create agent failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	var created map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if created["name"] != "test-agent" {
		t.Errorf("expected name 'test-agent', got '%v'", created["name"])
	}

	agentID := created["id"].(string)

	// List agents
	resp2, err := http.Get(baseURL + "/api/v1/agents")
	if err != nil {
		t.Fatalf("list agents failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp2.StatusCode)
	}

	// Delete agent
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/agents/%s", baseURL, agentID), nil)
	resp3, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("delete agent failed: %v", err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp3.StatusCode)
	}
}

func TestMemoryCRUD(t *testing.T) {
	// Create memory
	memory := map[string]interface{}{
		"type":    "semantic",
		"content": "Test memory content",
		"project": "test-project",
	}

	body, _ := json.Marshal(memory)
	resp, err := http.Post(baseURL+"/api/v1/memories", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create memory failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	var created map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if created["content"] != "Test memory content" {
		t.Errorf("expected content 'Test memory content', got '%v'", created["content"])
	}

	// Search memories
	resp2, err := http.Get(baseURL + "/api/v1/memories?query=test")
	if err != nil {
		t.Fatalf("search memories failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp2.StatusCode)
	}
}

func TestSessionLifecycle(t *testing.T) {
	// Start session
	session := map[string]interface{}{
		"project": "test-project",
		"task":    "integration test",
	}

	body, _ := json.Marshal(session)
	resp, err := http.Post(baseURL+"/api/v1/sessions", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("start session failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	var created map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if created["project"] != "test-project" {
		t.Errorf("expected project 'test-project', got '%v'", created["project"])
	}

	sessionID := created["id"].(string)

	// End session
	endBody, _ := json.Marshal(map[string]string{"summary": "Test completed"})
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/v1/sessions/%s/end", baseURL, sessionID), bytes.NewReader(endBody))
	req.Header.Set("Content-Type", "application/json")
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("end session failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp2.StatusCode)
	}
}
