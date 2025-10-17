# Gemini AI Assistant Guidelines for Flyt Project

## Understanding Flyt
Flyt is a lightweight Go workflow framework designed for building LLM-powered applications. It provides a node-based architecture where workflows are composed of interconnected nodes, each performing specific tasks.

## Quick Commands
- Build & Run: `go run . [-mode agent|batch]`
- Test all: `go test ./...`
- Test specific: `go test -run ^TestFunctionName$ ./...`
- Format: `go fmt ./...`
- Lint: `go vet ./...`
- Clean deps: `go mod tidy`

## Flyt Architecture

### Node Lifecycle
Each node can have three phases:
1. **PrepFunc**: Gather and validate data from SharedStore
2. **ExecFunc**: Perform the main task (required)
3. **PostFunc**: Store results and determine next action

```go
func CreateProcessingNode() flyt.Node {
    return flyt.NewNode(
        // Prep: Get data from shared store
        flyt.WithPrepFunc(func(ctx context.Context, shared *flyt.SharedStore) (any, error) {
            input, ok := shared.Get("input")
            if !ok {
                return nil, fmt.Errorf("missing input")
            }
            return input, nil
        }),
        // Exec: Process the data
        flyt.WithExecFunc(func(ctx context.Context, prepResult any) (any, error) {
            // Main processing logic here
            processed := processData(prepResult)
            return processed, nil
        }),
        // Post: Store result and route
        flyt.WithPostFunc(func(ctx context.Context, shared *flyt.SharedStore, prepResult, execResult any) (flyt.Action, error) {
            shared.Set("processed", execResult)
            if needsMoreProcessing(execResult) {
                return "continue", nil
            }
            return flyt.DefaultAction, nil
        }),
    )
}
```

### Flow Construction
Flows define how nodes connect and execute:
```go
flow := flyt.NewFlow(startNode)
// Connect nodes with actions
flow.Connect(startNode, "continue", processNode)
flow.Connect(startNode, "skip", endNode)
flow.Connect(processNode, flyt.DefaultAction, endNode)
```

### SharedStore Pattern
Data sharing between nodes:
```go
// Write data
shared.Set("key", value)
shared.Set("results", []string{"a", "b", "c"})

// Read data
value, exists := shared.Get("key")
if !exists {
    // Handle missing data
}

// Type assertion
results, ok := shared.Get("results").([]string)
```

## Key Patterns

### Error Handling with Retries
```go
flyt.NewNode(
    flyt.WithExecFunc(unreliableOperation),
    flyt.WithMaxRetries(3),
    flyt.WithRetryDelay(2*time.Second),
)
```

### Batch Processing
```go
flyt.NewBatchNode(
    flyt.WithBatchSize(20),
    flyt.WithBatchExecFunc(func(ctx context.Context, items []any) ([]any, error) {
        // Process items in parallel
        return processedItems, nil
    }),
)
```

### Conditional Routing
```go
flyt.WithPostFunc(func(ctx context.Context, shared *flyt.SharedStore, prepResult, execResult any) (flyt.Action, error) {
    result := execResult.(ProcessResult)
    switch result.Status {
    case "success":
        return "handleSuccess", nil
    case "retry":
        return "retryProcess", nil
    default:
        return "handleError", nil
    }
})
```

## Code Style
- Import order: stdlib → github.com/mark3labs/flyt → internal
- Always handle errors explicitly
- Use meaningful action names for flow control
- Keep nodes focused on single responsibilities
- Document exported functions and types

## Project Layout
- `main.go`: CLI entry, mode selection
- `nodes.go`: All node implementations
- `flow.go`: Flow definitions for different modes
- `utils/llm.go`: OpenAI integration
- `utils/search.go`: Web search utilities
- `utils/text.go`: Text processing helpers

## Environment Setup
- `GEMINI_API_KEY`: Required for LLM functionality
- Execution modes:
  - `agent`: Interactive Q&A mode
  - `batch`: Bulk processing mode