package chunker

import (
	"log"
	"strings"
	"unicode"
)

func ChunkText(content string, chunkSize int) ([]string, error) {
	log.Printf("Starting text chunking with chunk size %d words", chunkSize)

	content = strings.ReplaceAll(content, "\r\n", " ")
	content = strings.ReplaceAll(content, "\n", " ")

	sentences := splitIntoSentences(content)
	log.Printf("Split content into %d sentences", len(sentences))

	return createChunksFromSentences(sentences, chunkSize), nil
}

func splitIntoSentences(text string) []string {
	log.Printf("Splitting text into sentences, text length: %d characters", len(text))

	text = replaceAbbreviations(text)

	var sentences []string
	var currentSentence strings.Builder

	for i := range text {
		currentSentence.WriteByte(text[i])

		if (text[i] == '.' || text[i] == '!' || text[i] == '?') &&
			(i == len(text)-1 || unicode.IsSpace(rune(text[i+1]))) {

			sentence := strings.TrimSpace(currentSentence.String())
			if len(strings.Fields(sentence)) > 0 {
				sentences = append(sentences, sentence)
			}
			currentSentence.Reset()
		}
	}

	if currentSentence.Len() > 0 {
		sentence := strings.TrimSpace(currentSentence.String())
		if len(strings.Fields(sentence)) > 0 {
			sentences = append(sentences, sentence)
		}
	}

	log.Printf("Found %d sentences in text", len(sentences))
	return sentences
}

func createChunksFromSentences(sentences []string, targetChunkSize int) []string {
	var chunks []string
	var currentChunk strings.Builder
	currentWordCount := 0

	for i, sentence := range sentences {
		sentenceWords := len(strings.Fields(sentence))

		if currentWordCount > 0 && currentWordCount+sentenceWords > targetChunkSize {
			chunk := strings.TrimSpace(currentChunk.String())
			chunks = append(chunks, chunk)
			log.Printf("Created chunk with %d words", currentWordCount)

			currentChunk.Reset()
			currentWordCount = 0
		}

		currentChunk.WriteString(sentence + " ")
		currentWordCount += sentenceWords

		if i > 0 && i%100 == 0 {
			log.Printf("Processed %d/%d sentences", i, len(sentences))
		}
	}

	if currentChunk.Len() > 0 {
		chunk := strings.TrimSpace(currentChunk.String())
		chunks = append(chunks, chunk)
		log.Printf("Created final chunk with %d words", currentWordCount)
	}

	log.Printf("Created %d chunks from %d sentences", len(chunks), len(sentences))
	return chunks
}

func replaceAbbreviations(text string) string {
	abbreviations := []string{
		"Mr.", "Mrs.", "Ms.", "Dr.", "Prof.",
		"Inc.", "Ltd.", "Co.", "Corp.",
		"i.e.", "e.g.", "etc.",
		"vs.", "a.m.", "p.m.",
		"U.S.", "U.K.", "E.U.",
	}

	result := text
	for _, abbr := range abbreviations {
		placeholder := strings.ReplaceAll(abbr, ".", "·")
		result = strings.ReplaceAll(result, abbr, placeholder)
	}

	return result
}
