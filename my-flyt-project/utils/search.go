package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// SearchResult represents a single search result
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Snippet     string `json:"snippet"`
	Description string `json:"description"`
}

// SearchWeb performs a web search using DuckDuckGo API
// In production, you might want to use a proper search API like Brave Search or Google Custom Search
func SearchWeb(query string) ([]SearchResult, error) {
	// For demonstration, we'll use a mock implementation
	// In production, integrate with a real search API

	results := []SearchResult{
		{
			Title:       fmt.Sprintf("Search result 1 for: %s", query),
			URL:         "https://example.com/1",
			Snippet:     "This is a snippet of the first search result...",
			Description: "Detailed description of the first result",
		},
		{
			Title:       fmt.Sprintf("Search result 2 for: %s", query),
			URL:         "https://example.com/2",
			Snippet:     "This is a snippet of the second search result...",
			Description: "Detailed description of the second result",
		},
		{
			Title:       fmt.Sprintf("Search result 3 for: %s", query),
			URL:         "https://example.com/3",
			Snippet:     "This is a snippet of the third search result...",
			Description: "Detailed description of the third result",
		},
	}

	return results, nil
}

// SearchWebDuckDuckGo performs a real web search using DuckDuckGo Instant Answer API
// Note: This API is limited and may not return results for all queries
func SearchWebDuckDuckGo(query string) ([]SearchResult, error) {
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
		url.QueryEscape(query))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse DuckDuckGo response
	var ddgResponse struct {
		Abstract       string `json:"Abstract"`
		AbstractText   string `json:"AbstractText"`
		AbstractSource string `json:"AbstractSource"`
		AbstractURL    string `json:"AbstractURL"`
		RelatedTopics  []struct {
			Text     string            `json:"Text"`
			FirstURL string            `json:"FirstURL"`
			Icon     map[string]string `json:"Icon"`
		} `json:"RelatedTopics"`
	}

	if err := json.Unmarshal(body, &ddgResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var results []SearchResult

	// Add abstract if available
	if ddgResponse.Abstract != "" {
		results = append(results, SearchResult{
			Title:       ddgResponse.AbstractSource,
			URL:         ddgResponse.AbstractURL,
			Snippet:     ddgResponse.AbstractText,
			Description: ddgResponse.Abstract,
		})
	}

	// Add related topics
	for _, topic := range ddgResponse.RelatedTopics {
		if topic.Text != "" {
			results = append(results, SearchResult{
				Title:       "Related Topic",
				URL:         topic.FirstURL,
				Snippet:     topic.Text,
				Description: topic.Text,
			})
		}
	}

	return results, nil
}

// FormatSearchResults formats search results into a string
func FormatSearchResults(results []SearchResult) string {
	if len(results) == 0 {
		return "No search results found."
	}

	formatted := fmt.Sprintf("Found %d search results:\n\n", len(results))

	for i, result := range results {
		formatted += fmt.Sprintf("%d. %s\n", i+1, result.Title)
		formatted += fmt.Sprintf("   URL: %s\n", result.URL)
		formatted += fmt.Sprintf("   %s\n\n", result.Snippet)
	}

	return formatted
}
