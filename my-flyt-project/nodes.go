package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flyt-project-template/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mark3labs/flyt"
)

// make a struct of user and ai conversation
type Conversation struct {
	User string
	AI   any
}

type History struct {
	Conversations []Conversation
}

// getHistory reads history from shared store and normalizes it to History.
func getHistory(shared *flyt.SharedStore) History {
	raw, _ := shared.Get("history")
	switch v := raw.(type) {
	case History:
		return v
	case []Conversation:
		return History{Conversations: v}
	case nil:
		return History{}
	default:
		// Best-effort conversion from []interface{} with map[string]interface{}
		if s, ok := raw.([]interface{}); ok {
			convs := make([]Conversation, 0, len(s))
			for _, it := range s {
				if m, ok := it.(map[string]interface{}); ok {
					var c Conversation
					if user, ok := m["User"].(string); ok {
						c.User = user
					}
					if ai, ok := m["AI"]; ok {
						c.AI = ai
					}
					convs = append(convs, c)
				}
			}
			return History{Conversations: convs}
		}
		return History{}
	}
}

// saveHistory writes the History back into the shared store.
func saveHistory(shared *flyt.SharedStore, h History) {
	shared.Set("history", h)
}

// CreateGetQuestionNode creates a node that gets a question from user input
// func CreateGetQuestionNode() flyt.Node {
// 	return flyt.NewNode(
// 		flyt.WithExecFunc(func(ctx context.Context, prepResult any) (any, error) {
// 			// Get question from user input
// 			reader := bufio.NewReader(os.Stdin)
// 			fmt.Print("Enter your question: ")
// 			userQuestion, err := reader.ReadString('\n')
// 			if err != nil {
// 				return nil, err
// 			}
// 			return strings.TrimSpace(userQuestion), nil
// 		}),
// 		flyt.WithPostFunc(func(ctx context.Context, shared *flyt.SharedStore, prepResult, execResult any) (flyt.Action, error) {
// 			// Store the user's question
// 			shared.Set("question", execResult)
// 			return flyt.DefaultAction, nil
// 		}),
// 	)
// }

// CreateAnswerNode creates a node that generates an answer using LLM
func CreateAnswerNode() flyt.Node {
	return flyt.NewNode(
		flyt.WithPrepFunc(func(ctx context.Context, shared *flyt.SharedStore) (any, error) {
			// Read question from shared store
			question, ok := shared.Get("question")
			if !ok {
				return nil, fmt.Errorf("no question found in shared store")
			}

			// Use helper to normalize history
			h := getHistory(shared)

			return map[string]any{
				"question": question,
				"history":  h.Conversations,
			}, nil
		}),
		flyt.WithExecFunc(func(ctx context.Context, prepResult any) (any, error) {
			data := prepResult.(map[string]any)
			question := data["question"].(string)
			history := data["history"].([]Conversation)
			context := data["context"].(string)

			// Get API key from environment
			apiKey := os.Getenv("GEMINI_API_KEY")
			if apiKey == "" {
				return nil, fmt.Errorf("GEMINI_API_KEY not set")
			}

			// Call LLM to get the answer
			// Build prompt including a short serialized history if present
			prompt := fmt.Sprintf("Context: %s\nAnswer this question: %s", context, question)
			if len(history) > 0 {
				// Serialize recent history entries into a simple text block
				var b strings.Builder
				for i, c := range history {
					b.WriteString(fmt.Sprintf("%d. User: %s\n   AI: %v\n", i+1, c.User, c.AI))
				}
				prompt = fmt.Sprintf("Context: %s\nHistory:\n%s\nAnswer this question: %s", context, b.String(), question)
			}

			// Call LLM helper in utils
			response, err := utils.CallLLM(prompt)
			if err != nil {
				return nil, err
			}

			return response, nil
		}),
		flyt.WithPostFunc(func(ctx context.Context, shared *flyt.SharedStore, prepResult, execResult any) (flyt.Action, error) {
			// Store the answer and append to history using helpers
			shared.Set("answer", execResult)
			q, _ := shared.Get("question")
			conv := Conversation{User: q.(string), AI: execResult}

			h := getHistory(shared)
			h.Conversations = append(h.Conversations, conv)
			saveHistory(shared, h)

			return flyt.DefaultAction, nil
		}),
	)
}

// CreateAnalyzeNode creates a node that analyzes input and decides next action
func CreateAnalyzeNode() flyt.Node {
	return flyt.NewNode(
		flyt.WithPrepFunc(func(ctx context.Context, shared *flyt.SharedStore) (any, error) {
			question, ok := shared.Get("question")
			if !ok {
				return nil, fmt.Errorf("no question found in shared store")
			}
			searchResults, _ := shared.Get("search_results")

			return map[string]any{
				"question":       question,
				"search_results": searchResults,
			}, nil
		}), flyt.WithExecFunc(func(ctx context.Context, prepResult any) (any, error) {
			data := prepResult.(map[string]any)

			// Simple logic to decide next action
			// In a real implementation, this could use an LLM to make decisions
			if data["search_results"] == nil {
				// No search results yet, might need to search
				return "search", nil
			}
			// prompt := fmt.Sprintf("Answer this question: %s", question)
			// if data["context"] != nil {
			// 	prompt = fmt.Sprintf("Context: %s\n\nAnswer this question: %s", data["context"], question)
			// }

			// We have search results, process them
			return "process", nil
		}),
		flyt.WithPostFunc(func(ctx context.Context, shared *flyt.SharedStore, prepResult, execResult any) (flyt.Action, error) {
			action := execResult.(string)
			return flyt.Action(action), nil
		}),
	)
}

// CreateSearchNode creates a node that performs web search
// CreateSearchNode creates a node that performs a real web search using Tavily AI
func CreateSearchNode() flyt.Node {
	return flyt.NewNode(
		flyt.WithPrepFunc(func(ctx context.Context, shared *flyt.SharedStore) (any, error) {
			question, ok := shared.Get("question")
			if !ok {
				return nil, fmt.Errorf("no question found in shared store")
			}
			apiKey := os.Getenv("TAVILY_API_KEY")
			if apiKey == "" {
				return nil, fmt.Errorf("TAVILY_API_KEY environment variable not set")
			}

			fmt.Println("Using Tavily API Key:", apiKey)
			return map[string]string{
				"question": question.(string),
				"apiKey":   apiKey,
			}, nil
		}),
		flyt.WithExecFunc(func(ctx context.Context, prepResult any) (any, error) {
			data := prepResult.(map[string]string)
			question := data["question"]
			apiKey := data["apiKey"]

			fmt.Println("🔎 Performing web search...")

			// 1. Prepare the request body for Tavily API
			requestBody, err := json.Marshal(map[string]interface{}{
				"api_key":      apiKey,
				"query":        question,
				"search_depth": "basic",
				"max_results":  3, // Get top 3 results
			})
			if err != nil {
				return nil, fmt.Errorf("failed to marshal search request: %w", err)
			}

			// 2. Make the HTTP POST request
			resp, err := http.Post("https://api.tavily.com/search", "application/json", bytes.NewBuffer(requestBody))
			if err != nil {
				return nil, fmt.Errorf("failed to make search request: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read search response: %w", err)
			}
			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("search API request failed with status %d: %s", resp.StatusCode, string(body))
			}

			// 3. Parse the JSON response
			var searchResponse struct {
				Results []struct {
					Title   string `json:"title"`
					URL     string `json:"url"`
					Content string `json:"content"`
				} `json:"results"`
			}
			if err := json.Unmarshal(body, &searchResponse); err != nil {
				return nil, fmt.Errorf("failed to parse search response: %w", err)
			}

			if len(searchResponse.Results) == 0 {
				return "No relevant search results found.", nil
			}

			// 4. Format results into a single string for the next LLM call
			var resultsBuilder strings.Builder
			resultsBuilder.WriteString("Web search results:\n\n")
			for i, result := range searchResponse.Results {
				resultsBuilder.WriteString(fmt.Sprintf("Source %d: %s (%s)\nContent: %s\n\n", i+1, result.Title, result.URL, result.Content))
			}

			return resultsBuilder.String(), nil
		}),
		flyt.WithPostFunc(func(ctx context.Context, shared *flyt.SharedStore, prepResult, execResult any) (flyt.Action, error) {
			shared.Set("search_results", execResult)
			// Now that we have results, go back to the analyze node to decide the next step
			return "analyze", nil
		}),
	)
}

// CreateProcessNode creates a node that processes information
func CreateProcessNode() flyt.Node {
	return flyt.NewNode(
		flyt.WithPrepFunc(func(ctx context.Context, shared *flyt.SharedStore) (any, error) {
			question, _ := shared.Get("question")
			searchResults, _ := shared.Get("search_results")

			return map[string]any{
				"question":       question,
				"search_results": searchResults,
			}, nil
		}),
		flyt.WithExecFunc(func(ctx context.Context, prepResult any) (any, error) {
			data := prepResult.(map[string]any)
			// question := data["question"].(string)
			searchResults := data["search_results"].(string)

			// Build prompt to process search results
			// prompt := fmt.Sprintf("Using the following search results, provide a detailed answer to the question: %s\n\nSearch Results:\n%s", question, searchResults)

			// Call LLM helper in utils
			// response, err := utils.CallLLM(prompt)
			// if err != nil {
			// 	return nil, err
			// }

			// Process the search results
			// In a real implementation, this could extract key information,
			// summarize, or transform the data
			// _ = data // Will be used when processing is implemented
			// processed := "Processed information from search results"
			return searchResults, nil

		}), flyt.WithPostFunc(func(ctx context.Context, shared *flyt.SharedStore, prepResult, execResult any) (flyt.Action, error) {
			shared.Set("context", execResult)
			// q, _ := shared.Get("question")
			// conv := Conversation{User: q.(string), AI: execResult}

			// h := getHistory(shared)
			// h.Conversations = append(h.Conversations, conv)
			// saveHistory(shared, h)
			return flyt.DefaultAction, nil
		}),
	)
}

// CreateLoadItemsNode creates a node that loads items for batch processing
func CreateLoadItemsNode() flyt.Node {
	return flyt.NewNode(
		flyt.WithExecFunc(func(ctx context.Context, prepResult any) (any, error) {
			// Load items from a source (file, API, database, etc.)
			// For demo, create some sample items
			items := []string{
				"Item 1",
				"Item 2",
				"Item 3",
				"Item 4",
				"Item 5",
			}

			return items, nil
		}),
		flyt.WithPostFunc(func(ctx context.Context, shared *flyt.SharedStore, prepResult, execResult any) (flyt.Action, error) {
			shared.Set(flyt.KeyItems, execResult)
			return flyt.DefaultAction, nil
		}),
	)
}

// CreateBatchProcessNode creates a node that processes items in batch
func CreateBatchProcessNode() flyt.Node {
	processFunc := func(ctx context.Context, item any) (any, error) {
		// Process each item
		itemStr := item.(string)
		return fmt.Sprintf("Processed: %s", itemStr), nil
	}

	// Use Flyt's built-in batch node
	return flyt.NewBatchNode(processFunc, true) // true for concurrent processing
}

// CreateAggregateResultsNode creates a node that aggregates batch results
func CreateAggregateResultsNode() flyt.Node {
	return flyt.NewNode(
		flyt.WithPrepFunc(func(ctx context.Context, shared *flyt.SharedStore) (any, error) {
			results, ok := shared.Get(flyt.KeyResults)
			if !ok {
				return nil, fmt.Errorf("no results found")
			}
			return results, nil
		}),
		flyt.WithExecFunc(func(ctx context.Context, prepResult any) (any, error) {
			results := prepResult.([]any)

			// Aggregate results
			var aggregated strings.Builder
			aggregated.WriteString("Aggregated Results:\n")

			for i, result := range results {
				aggregated.WriteString(fmt.Sprintf("%d. %v\n", i+1, result))
			}

			return aggregated.String(), nil
		}),
		flyt.WithPostFunc(func(ctx context.Context, shared *flyt.SharedStore, prepResult, execResult any) (flyt.Action, error) {
			shared.Set("final_results", execResult)
			fmt.Println(execResult)
			return flyt.DefaultAction, nil
		}),
	)
}
