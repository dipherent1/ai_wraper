package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"flyt-project-template/utils"

	"github.com/joho/godotenv"
	"github.com/mark3labs/flyt"
)

func readMultiLineInput(reader *bufio.Reader) (string, error) {
	var builder strings.Builder
	fmt.Println("(Enter your text. Type EOF on a new line or press Ctrl+D to finish)")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// io.EOF is the signal sent by Ctrl+D. It's not a "real" error.
			if err == io.EOF {
				return builder.String(), nil
			}
			// A different, unexpected error occurred.
			return "", err
		}

		// Check if the user typed the delimiter.
		if strings.TrimSpace(line) == "EOF" {
			break
		}

		// Add the line to our builder.
		builder.WriteString(line)
	}

	return builder.String(), nil
}

func displayAnswer(answer string) error {
	// Create a secure temporary file to hold the markdown content.
	tmpFile, err := os.CreateTemp("", "ai-answer-*.md")
	if err != nil {
		return fmt.Errorf("could not create temp file: %w", err)
	}
	// IMPORTANT: Ensure the temporary file is deleted when we're done.
	defer os.Remove(tmpFile.Name())

	// Write the AI's answer into the temporary file.
	if _, err := tmpFile.Write([]byte(answer)); err != nil {
		return fmt.Errorf("could not write to temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("could not close temp file: %w", err)
	}

	// Prepare the 'glow' command to render the file.
	// We use '-p' to make it behave like a pager (like 'less').
	cmd := exec.Command("glow", tmpFile.Name())

	// Connect the command's output directly to your terminal's output.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command. This will take over the terminal to display the content.
	return cmd.Run()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	// Define command line flags
	var (
		mode          = flag.String("mode", "qa", "Flow mode: qa, agent, or batch")
		verbose       = flag.Bool("v", false, "Enable verbose output")
		model         = flag.String("model", "gemini-2.5-flash", "LLM model to use")
		imagePathsStr = flag.String("images", "", "Comma-separated list of image paths")
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
	var initialImagePaths []string
	if *imagePathsStr != "" {
		// Split the comma-separated string into a slice of paths
		initialImagePaths = strings.Split(*imagePathsStr, ",")
		fmt.Printf("🖼️ Loaded %d image(s) from command line.\n", len(initialImagePaths))
	}
	shared.Set("image_paths", initialImagePaths) // Set it once at the start

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
		fmt.Print("\nYou: ")
		// Call our new multi-line input function instead of the single-line read.
		userInput, err := readMultiLineInput(reader)
		if err != nil {
			log.Fatalf("Failed to read input: %v", err)
		}
		userInput = strings.TrimSpace(userInput)

		// If the user enters *only* "quit" or "exit", we should still quit.
		// If they enter nothing (just Ctrl+D on an empty prompt), we should prompt again.
		if userInput == "" {
			continue
		}
		if strings.ToLower(userInput) == "quit" || strings.ToLower(userInput) == "exit" {
			fmt.Println("🤖 Goodbye!")
			break
		}

		shared.Set("question", userInput)

		fmt.Println("🚀 Running flow...")
		err = flow.Run(ctx, shared)
		if err != nil {
			log.Fatalf("❌ Flow failed: %v", err)
		}

		fmt.Println("\n🎉 Flow completed successfully!")
		if answer, ok := shared.Get("answer"); ok {
			fmt.Println("\n✅ Answer:")
			fmt.Println(answer)
			if err := displayAnswer(answer.(string)); err != nil {
				// If Glow fails, fall back to plain text.
				fmt.Println("Glow renderer failed, printing raw text:")
				fmt.Println(answer)
			}
		}
	}

}
