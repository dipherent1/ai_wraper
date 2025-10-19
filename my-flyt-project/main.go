package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"flyt-project-template/utils"

	"github.com/joho/godotenv"
	"github.com/mark3labs/flyt"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	// Define command line flags
	var (
		mode    = flag.String("mode", "qa", "Flow mode: qa, agent, or batch")
		verbose = flag.Bool("v", false, "Enable verbose output")
		model   = flag.String("model", "gemini-2.5-flash", "LLM model to use")
	)
	// Parse flags first, then set package-level default model in utils so other packages use the selected model
	flag.Parse()
	utils.DefaultModel = *model
	log.Printf("Setting default LLM model to: %s", utils.DefaultModel)

	// Check for required environment variables
	if os.Getenv("GEMINI_API_KEY") == "" {
		log.Println("Warning: GEMINI_API_KEY not set. Some features may not work.")
	}

	// Create shared store
	shared := flyt.NewSharedStore()
	var history History
	// Store the full History struct (not just the slice) for easier retrieval
	shared.Set("history", history)
	shared.Set("context", " you are a helpful assistant. ")

	// Create context
	ctx := context.Background()

	// Select and run the appropriate flow
	var flow *flyt.Flow

	switch *mode {
	case "qa":
		fmt.Println("🤖 Starting Q&A Flow...")
		flow = CreateQAFlow()

	case "agent":
		fmt.Println("🤖 Starting Agent Flow...")
		flow = CreateAgentFlow()
		// For agent mode, we need to set an initial question

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

	reader := bufio.NewReader(os.Stdin)
	for {

		// Prompt for question if not provided
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

		fmt.Println("🚀 Running flow...")
		err = flow.Run(ctx, shared)
		if err != nil {
			log.Fatalf("❌ Flow failed: %v", err)
		}

		fmt.Println("\n🎉 Flow completed successfully!")
		if answer, ok := shared.Get("answer"); ok {
			fmt.Println("\n✅ Answer:")
			fmt.Println(answer)
		}
	}

	// Run the flow

	// Display results based on mode
	// switch *mode {
	// case "qa", "agent":

	// case "batch":
	// 	if results, ok := shared.Get("final_results"); ok {
	// 		fmt.Println("\n✅ Batch Processing Complete:")
	// 		fmt.Println(results)
	// 	}
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
