package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// LLMConfig holds configuration for LLM calls
type LLMConfig struct {
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

// DefaultLLMConfig returns default configuration for Gemini
func DefaultLLMConfig() *LLMConfig {

	// Use package-level DefaultModel if set, otherwise fallback to a sane default
	model := DefaultModel
	if model == "" {
		model = "gemini-2.5-flash"
	}
	log.Printf("Using LLM model: %s", model)

	return &LLMConfig{
		Model:       model,
		Temperature: 0.7,
		MaxTokens:   0, // Use model default
	}
}

// DefaultModel is the package-level model name used when creating default configs.
// It can be set by the application (for example in `main.go`) after parsing flags.
var DefaultModel string

// CallLLM calls the Gemini API with the given prompt
func CallLLM(prompt string) (string, error) {
	return CallLLMWithConfig(prompt, DefaultLLMConfig())
}

// CallLLMWithConfig calls the Gemini API with custom configuration
func CallLLMWithConfig(prompt string, config *LLMConfig) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	// Prepare request body for Gemini API
	requestBody := map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]any{
			"temperature": config.Temperature,
		},
	}

	// Add max_tokens to generationConfig if specified
	if config.MaxTokens > 0 {
		genConfig := requestBody["generationConfig"].(map[string]any)
		genConfig["maxOutputTokens"] = config.MaxTokens
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request for Gemini API
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", config.Model, apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make request with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Gemini API response
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

// CallLLMStreaming calls the Gemini API with streaming response
// This is useful for long responses where you want to show progress
func CallLLMStreaming(prompt string, onChunk func(string) error) error {
	// Implementation would handle streaming responses (e.g., using server-sent events)
	// For now, we'll use the regular call as in the original code
	response, err := CallLLM(prompt)
	if err != nil {
		return err
	}

	return onChunk(response)
}
