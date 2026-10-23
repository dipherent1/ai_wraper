package utils

import (
	"fmt"
	"strings"
	"unicode"
)

// TextOperation represents different text processing operations
type TextOperation string

const (
	OpSummarize TextOperation = "summarize"
	OpExtract   TextOperation = "extract"
	OpClean     TextOperation = "clean"
	OpTokenize  TextOperation = "tokenize"
)

// ProcessText performs various text processing operations
func ProcessText(text string, operation TextOperation) (string, error) {
	switch operation {
	case OpSummarize:
		return SummarizeText(text)
	case OpExtract:
		return ExtractKeyPoints(text)
	case OpClean:
		return CleanText(text)
	case OpTokenize:
		tokens := TokenizeText(text)
		return strings.Join(tokens, " "), nil
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// SummarizeText creates a summary of the input text
// In a real implementation, this would use an LLM
func SummarizeText(text string) (string, error) {
	if len(text) < 100 {
		return text, nil
	}

	// For demo purposes, return first 100 characters
	// In production, use CallLLM with a summarization prompt
	summary := text[:100] + "..."
	return summary, nil
}

// ExtractKeyPoints extracts key points from text
func ExtractKeyPoints(text string) (string, error) {
	// Simple implementation: extract sentences with key phrases
	sentences := strings.Split(text, ".")
	var keyPoints []string

	keyPhrases := []string{"important", "key", "main", "critical", "essential"}

	for _, sentence := range sentences {
		lower := strings.ToLower(sentence)
		for _, phrase := range keyPhrases {
			if strings.Contains(lower, phrase) {
				keyPoints = append(keyPoints, strings.TrimSpace(sentence))
				break
			}
		}
	}

	if len(keyPoints) == 0 {
		// If no key phrases found, return first sentence
		if len(sentences) > 0 {
			keyPoints = append(keyPoints, strings.TrimSpace(sentences[0]))
		}
	}

	return strings.Join(keyPoints, ". "), nil
}

// CleanText removes extra whitespace and normalizes text
func CleanText(text string) (string, error) {
	// Remove extra whitespace
	text = strings.TrimSpace(text)

	// Replace multiple spaces with single space
	text = strings.Join(strings.Fields(text), " ")

	// Remove non-printable characters
	var cleaned strings.Builder
	for _, r := range text {
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			cleaned.WriteRune(r)
		}
	}

	return cleaned.String(), nil
}

// TokenizeText splits text into tokens (words)
func TokenizeText(text string) []string {
	// Simple word tokenization
	// In production, you might use a more sophisticated tokenizer

	// Convert to lowercase and split by whitespace
	text = strings.ToLower(text)
	words := strings.Fields(text)

	// Clean punctuation from words
	var tokens []string
	for _, word := range words {
		// Remove leading and trailing punctuation
		word = strings.TrimFunc(word, func(r rune) bool {
			return unicode.IsPunct(r)
		})

		if word != "" {
			tokens = append(tokens, word)
		}
	}

	return tokens
}

// ChunkText splits text into chunks of specified size
func ChunkText(text string, chunkSize int) []string {
	if chunkSize <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	var chunks []string
	var currentChunk []string
	currentSize := 0

	for _, word := range words {
		wordLen := len(word) + 1 // +1 for space

		if currentSize+wordLen > chunkSize && len(currentChunk) > 0 {
			// Start new chunk
			chunks = append(chunks, strings.Join(currentChunk, " "))
			currentChunk = []string{word}
			currentSize = wordLen
		} else {
			currentChunk = append(currentChunk, word)
			currentSize += wordLen
		}
	}

	// Add last chunk
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.Join(currentChunk, " "))
	}

	return chunks
}

// CountTokens estimates the number of tokens in text
// This is a simple approximation - for accurate counts use a proper tokenizer
func CountTokens(text string) int {
	// Rough estimate: 1 token ≈ 4 characters or 0.75 words
	words := len(strings.Fields(text))
	chars := len(text)

	// Use the more conservative estimate
	tokensByWords := int(float64(words) / 0.75)
	tokensByChars := chars / 4

	if tokensByWords > tokensByChars {
		return tokensByWords
	}
	return tokensByChars
}
