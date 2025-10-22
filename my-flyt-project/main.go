package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"flyt-project-template/utils"

	"github.com/joho/godotenv"
	"github.com/mark3labs/flyt"
)

var ConversationName string

func TruncateString(s string, n int) string {
	// If the string has N or fewer characters, return the whole string.
	if utf8.RuneCountInString(s) <= n {
		return s
	}

	// Otherwise, convert to a slice of runes, take the first N, and convert back.
	runes := []rune(s)
	return string(runes[0:n])
}

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
	tmpFile, err := os.CreateTemp("", "ai-answer-*.md")
	if err != nil {
		return fmt.Errorf("could not create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(answer)); err != nil {
		return fmt.Errorf("could not write to temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("could not close temp file: %w", err)
	}

	// --- THIS IS THE ONLY LINE THAT CHANGES ---
	// We use 'bat' with flags for a clean, non-interactive output.
	cmd := exec.Command("bat", "--paging=never", "--style=plain", "--language=markdown", tmpFile.Name())
	// ------------------------------------------

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func setupSignalHandler(shared *flyt.SharedStore) {
	// Create a channel to receive OS signals.
	sigChan := make(chan os.Signal, 1)

	// Tell the OS to notify our channel when an interrupt (Ctrl+C) or terminate signal occurs.
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start a new goroutine. This will run in the background without blocking the main chat loop.
	go func() {
		// This line will block until a signal is received on the channel.
		<-sigChan

		// Once the signal is caught, we start the shutdown procedure.
		fmt.Println("\n🤖 Interrupt signal received. Saving conversation...")

		// Use the existing GetHistory helper to retrieve the latest conversation data.
		// NOTE: You must copy your 'GetHistory' function from nodes.go into main.go
		// or move it to a shared 'utils' package to make it accessible here.
		history := utils.GetHistory(shared)

		// If there's nothing to save, just exit.
		if len(history.Conversations) == 0 {
			fmt.Println("No conversation to save. Exiting.")
			os.Exit(0)
		}

		// Marshal the history struct into a nicely formatted JSON.
		jsonData, err := json.MarshalIndent(history, "", "  ")
		if err != nil {
			log.Printf("Error marshalling history to JSON: %v", err)
			os.Exit(1) // Exit with an error code
		}

		// Ensure the Conversations directory exists.
		dir := "Conversations"
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Printf("Error creating directory %s: %v", dir, err)
			os.Exit(1)
		}

		// Create a unique filename with a timestamp.
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		baseName := timestamp
		if ConversationName != "" {
			// sanitize spaces for filename
			baseName = strings.ReplaceAll(ConversationName, " ", "_") + "_" + timestamp
		}
		fileName := dir + string(os.PathSeparator) + baseName + ".json"

		// Write the JSON data to the file.
		err = os.WriteFile(fileName, jsonData, 0644)
		if err != nil {
			log.Printf("Error writing conversation to file: %v", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Conversation successfully saved to %s\n", fileName)
		os.Exit(0) // Exit the program cleanly
	}()
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
	var history utils.History
	// Store the full History struct (not just the slice) for easier retrieval
	shared.Set("history", history)
	setupSignalHandler(shared)

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
		if ConversationName == "" {
			ConversationName = TruncateString(userInput, 20)
			ConversationName = strings.ReplaceAll(ConversationName, " ", "_")
			shared.Set("conversation_name", ConversationName)

		}

		fmt.Println("🚀 Running flow...")
		err = flow.Run(ctx, shared)
		if err != nil {
			log.Fatalf("❌ Flow failed: %v", err)
		}

		fmt.Println("\n🎉 Flow completed successfully!")
		if answer, ok := shared.Get("answer"); ok {
			fmt.Println("\n✅ Answer:")
			// fmt.Println(answer)
			if err := displayAnswer(answer.(string)); err != nil {
				// If Glow fails, fall back to plain text.
				fmt.Println("Glow renderer failed, printing raw text:")
				fmt.Println(answer)
			}
		}
	}

}
