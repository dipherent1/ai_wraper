# Flyt Project Guidelines for AI Agents

## Build/Test Commands
- Run: `go run . [-mode agent|batch]`
- Test all: `go test ./...`
- Test single: `go test -run ^TestName$ ./...`
- Format: `go fmt ./...`
- Lint: `go vet ./...`
- Dependencies: `go mod tidy`

## Flyt Core Concepts
- **Nodes**: Units of work with Prep→Exec→Post lifecycle
- **Flows**: Connect nodes via actions to create workflows
- **SharedStore**: Pass data between nodes with Set/Get
- **Actions**: Control flow routing (DefaultAction or custom)

## Code Patterns
```go
// Node: flyt.NewNode(options...)
WithPrepFunc(ctx, shared) → prepare data
WithExecFunc(ctx, prepResult) → main logic (required)
WithPostFunc(ctx, shared, prep, exec) → store & route

// Flow: Connect nodes with actions
flow.Connect(fromNode, "action", toNode)
```

## Style Guide
- Imports: stdlib → github.com/mark3labs/flyt → internal
- Errors: lowercase, wrap with context
- Naming: unexported=camelCase, exported=PascalCase
- Context: always check cancellation
- Environment: `GEMINI_API_KEY` for LLM features