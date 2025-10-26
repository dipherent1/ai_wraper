# Flyt Project Template

This is a modification of A minimalist workflow template for building LLM applications with [Flyt](https://github.com/mark3labs/flyt), a Go-based workflow framework with zero dependencies.

I made this for my personal use to configure my fedora-hyprland and have a fast answering agent powered by Gemini API to avoid having to open a browser for simple questions.
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
git clone <repo-url>
cd flyt-project-template
```

2. Install dependencies:
```bash
go mod tidy
sudo dnf install bat #for displaying in terminal    

```

3. Set your API key:
```bash
export GEMINI_API_KEY="your-api-key-here"
export SERPAPI_API_KEY="your-api-key-here" #optional its not implemented yet or called in the project
export SYSTEM_INSTRUCTIONS_PATH="Your path here"
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
## my-flyt-project — Flyt LLM workflow starter

A minimalist Go template that demonstrates building LLM-powered workflows using the Flyt graph-based workflow pattern.

This repository contains a small example app that wires together modular "nodes" (prep/exec/post phases) and a lightweight `utils/llm.go` helper to call the Google Gemini API (the code is also easy to adapt for other providers).

Highlights

- Flow-based architecture using reusable nodes
- LLM utilities supporting text, search-enabled calls, images, and streaming helper
- Small, dependency-light Go app intended as a starting point for experiments and integrations

## Quick start

Prerequisites

- Go 1.21 or later
- A Gemini API key set in the `GEMINI_API_KEY` environment variable (see "Configuration" below)

Clone and prepare

```bash
git clone <repo-url>
cd my-flyt-project
go mod tidy
sudo dnf install bat #for displaying in terminal    

```

Build

```bash
go build -o build_files/ai-query
```

Run

```bash
# set API key (example)
export GEMINI_API_KEY="your-api-key-here"
export SYSTEM_INSTRUCTIONS_PATH="Your path here"

# run the binary
./build_files/ai-query
```

You can also run in-place during development:

```bash
GEMINI_API_KEY="your-api-key" 
SYSTEM_INSTRUCTIONS_PATH="Your path here"

go run .

```

## Configuration

Environment variables used by the project

- GEMINI_API_KEY (required): API key used by `utils/llm.go` to call Google's Generative Language API.
- SYSTEM_INSTRUCTIONS_PATH (optional): Path to a markdown file with system instructions. Defaults to `config/system_instructions.md`.

Runtime configuration in code

- The package-level variable `utils.DefaultModel` may be set by the application (for example in `main.go`) to override the default model (`gemini-2.5-flash`).
- `LLMConfig` controls temperature and optionally `MaxTokens`.

System instructions

- The file `config/system_instructions.md` contains the persistent system prompt. You can customize it to set tone, safety rules, and formatting preferences. The contents are sent to the model via the `systemInstruction` field on each request.
- To use a different file per environment or run, set `SYSTEM_INSTRUCTIONS_PATH`.

## LLM Utilities (what's in `utils/llm.go`)

The helper exposes several convenience functions:

- CallLLM(prompt string) (string, error): Simple text-only call using default config.
- CallLLMWithSearch(prompt string) (string, error): Enables the search tool in the request so the model can ground answers with web sources; returned text will include a **Sources** section if grounding data is present.
- CallLLMWithImages(prompt string, imagePaths []string) (string, error): Send images alongside a text prompt by base64-encoding image files and attaching them to the request.
- CallLLMWithConfig(prompt string, config *LLMConfig, useSearch bool) (string, error): Lower-level call that accepts config and an indicator to enable search tools.
- CallLLMStreaming(...): A placeholder wrapper that currently calls the non-streaming call and forwards chunks — useful future extension.

Notes on behavior

- If `useSearch` is true, the request contains a `tools` section which causes Gemini to return grounding metadata (sources). The helper formats sources into a markdown list under `---\n**Sources:**`.
- Images are encoded in base64 and a MIME type is inferred from the file extension (.jpg, .png, .webp, .heic, .heif). Unsupported extensions return an error.

## Project layout (important files)

- `main.go` — application entrypoint. Parses flags, sets `utils.DefaultModel` when needed and runs the flow.
- `flow.go` — builds and connects your nodes into a Flyt flow.
- `nodes.go` — node implementations (prep/exec/post lifecycle).
- `utils/llm.go` — LLM helper functions (Gemini integration, images, search, streaming helper).
- `docs/design.md` — higher-level design notes and architecture rationale.
- `build_files/` — helper folder for output binaries and assets (example: `image.png`).

If you add nodes, keep them small and single-purpose. Each node should read from the shared store in prep, perform main work in exec, and write back in post.

## Examples

Basic question/answer flow (conceptual)

1. Start node reads input text from the shared store.
2. LLM node calls `utils.CallLLMWithSearch` to answer and attach grounding metadata.
3. Post node stores the answer and either ends the flow or routes to follow-up nodes.

Image prompts

1. Provide image paths (e.g., `build_files/image.png`) to `CallLLMWithImages` to include images in the prompt.

Batching and concurrency

If you implement batch nodes, make them idempotent and control concurrency from the node implementation.

## Development notes and edge cases

Edge cases to consider when extending this project

- Missing API key: `utils.getGEMINIAPIKey()` will return an error if `GEMINI_API_KEY` is not set.
- Large responses or long-running ops: use context with timeouts or the streaming helper to limit memory usage.
- Unsupported image formats: `CallLLMWithImages` will error for unknown extensions.
- Rate limits & retries: add retry and exponential backoff around LLM calls if you expect network flakiness.

## Tests and quality gates

There are no unit tests included by default. When adding code, aim to cover:

- Node prep/exec/post logic (small unit tests)
- LLM helper serialization and error handling (mock HTTP responses)

## Contributing

1. Fork the repo
2. Create a feature branch
3. Add tests for new behavior
4. Run `go vet` and `go test ./...`
5. Open a PR describing the changes

## License

This project is MIT licensed. See the `LICENSE` file in the repository root for details.

## Where to go next

- Read `docs/design.md` for architecture intent
- Inspect `nodes.go` and `flow.go` to understand the sample flow
- Try changing `utils.DefaultModel` in `main.go` to experiment with different Gemini models

---

If you want, I can also add a short example section in `main.go` showing how to set `utils.DefaultModel` from flags and a ready-to-run sample flow.

## Hyprland / Rofi launcher (ask-ai script)

There is a small launcher script included (at the repository root: `ask-ai`) that I use with Hyprland and Rofi to quickly call the compiled AI binary from a keyboard-driven menu. The script is a lightweight wrapper that:

- Presents a Rofi menu for common actions (quick question, deeper analysis, clipboard image, screenshot region, full-screen screenshot)
- Lets you pick a Gemini model (example values: `gemini-2.5-pro`, `gemini-2.5-flash`)
- Passes mode, model and optional `-images` arguments to your compiled binary

Important variables in the script

- `PROJECT_DIR` — set this to the full path of your `my-flyt-project` folder (where the binary and `build_files/` live).
- `AI_BINARY` — the name of the compiled binary inside the project (for example `build_files/ai-query` or simply `ai-query`).

Example install (copy the script to `~/bin` and make executable):

```bash
# from the project root
cp ask-ai ~/bin/ask-ai
chmod +x ~/bin/ask-ai

# Edit the script and set PROJECT_DIR and AI_BINARY, or wrap with a small launcher that passes them.
```

Usage in Hyprland/rofi

- Bind a key in your Hyprland config to execute `~/bin/ask-ai` (or call it from a panel/launcher). The script will open a Rofi menu and launch the AI binary inside a terminal emulator.

Notes

- The `ask-ai` script uses `grim`/`slurp`/`wl-paste` on Wayland for screenshots and clipboard handling; adjust commands if you use X11 or different tools.
- When passing screenshot/clipboard images, the script creates a temporary PNG and cleans it up after the binary finishes.

If you'd like, I can add a short example Hyprland keybind snippet and a tiny systemd user service to make the launcher available system-wide.
