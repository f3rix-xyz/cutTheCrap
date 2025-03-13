package chunker

import (
	"strings"
	"unicode"
)

func ChunkText(content string, chunkSize int) ([]string, error) {
	var chunks []string
	var currentChunk strings.Builder
	currentWordCount := 0

	paragraphs := splitParagraphs(content)

	for _, para := range paragraphs {
		words := strings.Fields(para)
		if currentWordCount+len(words) > chunkSize && currentWordCount > 0 {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
			currentWordCount = 0
		}
		currentChunk.WriteString(para + "\n\n")
		currentWordCount += len(words)
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}
	return chunks, nil
}

func splitParagraphs(text string) []string {
	split := strings.Split(text, "\n\n")
	var filtered []string
	for _, p := range split {
		if isParagraph(p) {
			filtered = append(filtered, strings.TrimSpace(p))
		}
	}
	return filtered
}

func isParagraph(s string) bool {
	return len(strings.Fields(s)) > 1 || strings.IndexFunc(s, unicode.IsPunct) != -1
}
