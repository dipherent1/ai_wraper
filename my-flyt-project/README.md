# Flyt Project Template

A minimalist workflow template for building LLM applications with [Flyt](https://github.com/mark3labs/flyt), a Go-based workflow framework with zero dependencies.

## Overview

This template provides a starting point for building LLM-powered applications using Flyt's graph-based workflow system. It includes:

- 📊 **Flow-based Architecture**: Model your LLM workflows as directed graphs
- 🔄 **Reusable Nodes**: Build modular components that handle specific tasks
- 🛡️ **Error Handling**: Built-in retry logic and fallback mechanisms
- 🚀 **Zero Dependencies**: Pure Go implementation for maximum portability

## Project Structure

```
flyt-project-template/
├── README.md           # This file
├── flow.go            # Flow definition and connections
├── main.go            # Application entry point
├── nodes.go           # Node implementations
├── go.mod             # Go module definition
├── docs/
│   └── design.md      # Design documentation
└── utils/
    ├── llm.go         # LLM integration utilities
    └── helpers.go     # General helper functions
```

## Quick Start

### Prerequisites

- Go 1.21 or later
- OpenAI API key (or other LLM provider)

### Setup

1. Clone this template:
```bash
git clone <your-repo-url>
cd flyt-project-template
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set your API key:
```bash
export GEMINI_API_KEY="your-api-key-here"
```

4. Run the example:
```bash
go run .
```

## Core Concepts

### Nodes

Nodes are the building blocks of your workflow. Each node has three phases:

1. **Prep** - Read from shared store and prepare data
2. **Exec** - Execute main logic (can be retried)
3. **Post** - Process results and decide next action

```go
node := flyt.NewNode(
    flyt.WithPrepFunc(func(ctx context.Context, shared *flyt.SharedStore) (any, error) {
        // Prepare data
        return data, nil
    }),
    flyt.WithExecFunc(func(ctx context.Context, prepResult any) (any, error) {
        // Execute logic
        return result, nil
    }),
    flyt.WithPostFunc(func(ctx context.Context, shared *flyt.SharedStore, prepResult, execResult any) (flyt.Action, error) {
        // Store results and return next action
        return flyt.DefaultAction, nil
    }),
)
```

### Flows

Flows connect nodes to create workflows:

```go
flow := flyt.NewFlow(startNode)
flow.Connect(startNode, "success", processNode)
flow.Connect(startNode, "error", errorNode)
flow.Connect(processNode, flyt.DefaultAction, endNode)
```

### Shared Store

Thread-safe data sharing between nodes:

```go
shared := flyt.NewSharedStore()
shared.Set("input", "Hello, Flyt!")
value, ok := shared.Get("input")
```

## Example Workflows

### Simple Q&A Flow

```go
// Create nodes
questionNode := CreateQuestionNode()
answerNode := CreateAnswerNode(apiKey)

// Connect nodes
flow := flyt.NewFlow(questionNode)
flow.Connect(questionNode, flyt.DefaultAction, answerNode)

// Run flow
shared := flyt.NewSharedStore()
err := flow.Run(context.Background(), shared)
```

### Agent with Decision Making

```go
// Create nodes with conditional routing
decideNode := CreateDecisionNode()
searchNode := CreateSearchNode()
answerNode := CreateAnswerNode()

// Build flow with branching
flow := flyt.NewFlow(decideNode)
flow.Connect(decideNode, "search", searchNode)
flow.Connect(decideNode, "answer", answerNode)
flow.Connect(searchNode, "decide", decideNode) // Loop back
```

## Advanced Features

### Batch Processing

Process multiple items concurrently:

```go
processFunc := func(ctx context.Context, item any) (any, error) {
    // Process each item
    return processItem(item), nil
}

batchNode := flyt.NewBatchNode(processFunc, true) // true for concurrent
```

### Error Handling & Retries

Add retry logic to handle transient failures:

```go
node := flyt.NewNode(
    flyt.WithExecFunc(func(ctx context.Context, prepResult any) (any, error) {
        return callFlakeyAPI()
    }),
    flyt.WithMaxRetries(3),
    flyt.WithWait(time.Second),
)
```

## Customization

### Adding New Nodes

1. Create a new node in `nodes.go`:
```go
func CreateMyCustomNode() flyt.Node {
    return flyt.NewNode(
        // Your implementation
    )
}
```

2. Add it to your flow in `flow.go`:
```go
customNode := CreateMyCustomNode()
flow.Connect(previousNode, "custom", customNode)
```

### Integrating Different LLMs

Modify `utils/llm.go` to support your preferred LLM provider:
- OpenAI
- Anthropic Claude
- Google Gemini
- Local models (Ollama, etc.)

## Best Practices

1. **Single Responsibility**: Each node should do one thing well
2. **Idempotency**: Nodes should be idempotent when possible
3. **Error Handling**: Always handle errors appropriately
4. **Context Awareness**: Respect context cancellation
5. **Logging**: Add appropriate logging for debugging

## Examples

Check out the [Flyt cookbook](https://github.com/mark3labs/flyt/tree/main/cookbook) for more examples:
- [Agent](https://github.com/mark3labs/flyt/tree/main/cookbook/agent) - AI agent with web search
- [Chat](https://github.com/mark3labs/flyt/tree/main/cookbook/chat) - Interactive chat application
- [MCP](https://github.com/mark3labs/flyt/tree/main/cookbook/mcp) - Model Context Protocol integration
- [Summarize](https://github.com/mark3labs/flyt/tree/main/cookbook/summarize) - Text summarization with retries

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This template is MIT licensed. See LICENSE file for details.

## Resources

- [Flyt Documentation](https://github.com/mark3labs/flyt)
- [Go Documentation](https://go.dev/doc/)
- [OpenAI API Reference](https://platform.openai.com/docs/api-reference)