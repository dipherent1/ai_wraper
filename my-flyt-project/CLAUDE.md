# Claude AI Assistant Guidelines for Flyt Project

## What is Flyt?
Flyt is a Go workflow framework for building LLM applications with zero external dependencies. It uses a node-based architecture where each node performs a specific task, and flows connect nodes to create complex workflows.

## Build/Test Commands
- Run app: `go run .` or `go run . -mode agent|batch`
- Run all tests: `go test ./...`
- Run single test: `go test -run ^TestName$ ./...`
- Format: `go fmt ./...`
- Lint: `go vet ./...`
- Dependencies: `go mod tidy`

## Core Concepts

### Nodes
Nodes are the basic units of work. Each node has three optional phases:
```go
node := flyt.NewNode(
    flyt.WithPrepFunc(func(ctx context.Context, shared *flyt.SharedStore) (any, error) {
        // Prepare data before execution
        return preparedData, nil
    }),
    flyt.WithExecFunc(func(ctx context.Context, prepResult any) (any, error) {
        // Main logic - this is the only required function
        return result, nil
    }),
    flyt.WithPostFunc(func(ctx context.Context, shared *flyt.SharedStore, prepResult, execResult any) (flyt.Action, error) {
        // Store results and decide next action
        shared.Set("key", execResult)
        return "nextAction", nil // or flyt.DefaultAction
    }),
)
```

### Flows
Flows connect nodes and define execution paths:
```go
flow := flyt.NewFlow(startNode)
flow.Connect(startNode, "success", successNode)
flow.Connect(startNode, "error", errorNode)
flow.Connect(successNode, flyt.DefaultAction, endNode)
```

### SharedStore
Pass data between nodes using the SharedStore:
```go
// In one node's PostFunc
shared.Set("userQuestion", question)

// In another node's PrepFunc
question, ok := shared.Get("userQuestion")
if !ok {
    return nil, fmt.Errorf("no question found")
}
```

### Actions
Actions determine flow routing. Return different actions from PostFunc to control flow:
- `flyt.DefaultAction`: Continue to default next node
- Custom actions: Route to specific nodes based on logic
- `flyt.StopAction`: Stop flow execution

## Code Style
- Imports: stdlib → github.com/mark3labs/flyt → internal packages
- Naming: unexported=camelCase, exported=PascalCase
- Error handling: Always check and wrap with context
- Documentation: Package comments + exported functions
- Node functions: Keep focused on single responsibility

## Common Patterns

### Retry Logic
```go
node := flyt.NewNode(
    flyt.WithExecFunc(apiCall),
    flyt.WithMaxRetries(3),
    flyt.WithRetryDelay(time.Second),
)
```

### Batch Processing
```go
batchNode := flyt.NewBatchNode(
    flyt.WithBatchSize(10),
    flyt.WithBatchExecFunc(func(ctx context.Context, items []any) ([]any, error) {
        // Process multiple items in parallel
    }),
)
```

### Context Handling
Always respect context cancellation:
```go
select {
case <-ctx.Done():
    return nil, ctx.Err()
default:
    // Continue processing
}
```

## Project Structure
- `main.go`: Entry point, mode handling
- `nodes.go`: Node implementations
- `flow.go`: Flow definitions
- `utils/`: Utilities (LLM, search, text processing)

## Environment
- `GEMINI_API_KEY`: Required for LLM features
- Modes: agent (interactive), batch (bulk processing)