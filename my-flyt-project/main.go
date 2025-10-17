package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mark3labs/flyt"
)

func main() {
	// Define command line flags
	var (
		mode    = flag.String("mode", "qa", "Flow mode: qa, agent, or batch")
		verbose = flag.Bool("v", false, "Enable verbose output")
	)
	flag.Parse()

	// Check for required environment variables
	if os.Getenv("GEMINI_API_KEY") == "" {
		log.Println("Warning: GEMINI_API_KEY not set. Some features may not work.")
	}

	// Create shared store
	shared := flyt.NewSharedStore()

	// Create context
	ctx := context.Background()

	// Select and run the appropriate flow
	var flow *flyt.Flow
	var err error

	switch *mode {
	case "qa":
		fmt.Println("🤖 Starting Q&A Flow...")
		flow = CreateQAFlow()

	case "agent":
		fmt.Println("🤖 Starting Agent Flow...")
		flow = CreateAgentFlow()
		// For agent mode, we need to set an initial question
		if flag.NArg() > 0 {
			shared.Set("question", flag.Arg(0))
		} else {
			// Prompt for question if not provided
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter your question: ")
			question, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("Failed to read input: %v", err)
			}
			question = strings.TrimSpace(question)
			if question == "" {
				question = "What is the capital of France?"
			}
			shared.Set("question", question)
		}

	case "batch":
		fmt.Println("🤖 Starting Batch Processing Flow...")
		flow = CreateBatchFlow()

	default:
		log.Fatalf("Unknown mode: %s. Use 'qa', 'agent', or 'batch'", *mode)
	}

	// Enable verbose logging if requested
	if *verbose {
		fmt.Println("📊 Verbose mode enabled")
		// In a real implementation, you might configure logging here
	}

	// Run the flow
	fmt.Println("🚀 Running flow...")
	err = flow.Run(ctx, shared)
	if err != nil {
		log.Fatalf("❌ Flow failed: %v", err)
	}

	// Display results based on mode
	switch *mode {
	case "qa", "agent":
		if answer, ok := shared.Get("answer"); ok {
			fmt.Println("\n✅ Answer:")
			fmt.Println(answer)
		}

	case "batch":
		if results, ok := shared.Get("final_results"); ok {
			fmt.Println("\n✅ Batch Processing Complete:")
			fmt.Println(results)
		}
	}

	fmt.Println("\n🎉 Flow completed successfully!")
}

// Example of how to run the application:
//
// Basic Q&A mode:
//   go run .
//
// Agent mode with a question:
//   go run . -mode agent "What is the capital of France?"
//
// Batch processing mode:
//   go run . -mode batch
//
// With verbose output:
//   go run . -v -mode qa
