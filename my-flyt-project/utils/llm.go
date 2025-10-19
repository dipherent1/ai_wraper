package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LLMConfig holds configuration for LLM calls
type LLMConfig struct {
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

type GroundingChunk struct {
	Web struct {
		URI   string `json:"uri"`
		Title string `json:"title"`
	} `json:"web"`
}

type GroundingMetadata struct {
	GroundingChunks []GroundingChunk `json:"groundingChunks"`
}

func getGEMINIAPIKey() (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}
	return apiKey, nil
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
	return CallLLMWithConfig(prompt, DefaultLLMConfig(), false) // 'false' for useSearch
}

func CallLLMWithSearch(prompt string) (string, error) {
	return CallLLMWithConfig(prompt, DefaultLLMConfig(), true) // 'true' for useSearch
}

func CallLLMWithConfig(prompt string, config *LLMConfig, useSearch bool) (string, error) {
	var builder strings.Builder
	builder.WriteString(prompt)
	builder.WriteString("\n always answer using markdown format.")
	prompt = builder.String()

	apiKey, err := getGEMINIAPIKey()
	if err != nil {
		return "", err
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

	// THE KEY CHANGE: If useSearch is true, add the "tools" section to the request
	if useSearch {
		requestBody["tools"] = []map[string]any{
			{
				"google_search": map[string]any{}, // This enables the tool
			},
		}
	}

	if config.MaxTokens > 0 {
		genConfig := requestBody["generationConfig"].(map[string]any)
		genConfig["maxOutputTokens"] = config.MaxTokens
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", config.Model, apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 60 * time.Second, // Increased timeout for potential search
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			GroundingMetadata GroundingMetadata `json:"groundingMetadata"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	answerText := result.Candidates[0].Content.Parts[0].Text

	if len(result.Candidates[0].GroundingMetadata.GroundingChunks) > 0 {
		var builder strings.Builder
		builder.WriteString(answerText) // Start with the answer
		builder.WriteString("\n\n---\n**Sources:**\n")

		// Loop through the sources and format them
		for i, chunk := range result.Candidates[0].GroundingMetadata.GroundingChunks {
			builder.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, chunk.Web.Title, chunk.Web.URI))
		}
		return builder.String(), nil
	}
	return answerText, nil

}

func CallLLMWithImages(prompt string, imagePaths []string) (string, error) {
	apiKey, err := getGEMINIAPIKey()
	if err != nil {
		return "", err
	}

	config := DefaultLLMConfig()

	// The key new logic starts here: we build a "parts" array containing
	// the text and all the encoded images.
	parts := []map[string]any{
		{"text": prompt}, // Start with the text prompt
	}

	for _, path := range imagePaths {
		// 1. Read the raw image file data
		imageData, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read image file %s: %w", path, err)
		}

		// 2. Base64 encode the image data
		encodedString := base64.StdEncoding.EncodeToString(imageData)

		// 3. Determine the MIME type from the file extension
		mimeType := ""
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		case ".webp":
			mimeType = "image/webp"
		case ".heic":
			mimeType = "image/heic"
		case ".heif":
			mimeType = "image/heif"
		default:
			return "", fmt.Errorf("unsupported image type: %s", ext)
		}

		// 4. Create the image part structure for the JSON request
		imagePart := map[string]any{
			"inline_data": map[string]any{
				"mime_type": mimeType,
				"data":      encodedString,
			},
		}
		parts = append(parts, imagePart)
	}

	// Now we build the final request body with our multi-part content
	requestBody := map[string]any{
		"contents": []map[string]any{
			{
				"role":  "user",
				"parts": parts, // Use the parts array we just built
			},
		},
		"generationConfig": map[string]any{
			"temperature": config.Temperature,
		},
	}
	// ... (The rest of the function is standard HTTP request logic, similar to before) ...
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", config.Model, apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 90 * time.Second} // Increased timeout for image uploads

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

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
