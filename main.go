package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"pdf-processor/internal/chunker"
	"pdf-processor/internal/config"
	"pdf-processor/internal/pdf"
	"pdf-processor/internal/workers"
	"strings"
	"time"
)

func main() {
	log.Println("Starting PDF processor service")
	cfg := config.Load()
	log.Printf("Configuration loaded: Port=%s, MaxConcurrent=%d, ChunkSize=%d", cfg.Port, cfg.MaxConcurrent, cfg.ChunkSize)

	http.HandleFunc("/process", uploadHandler(cfg))
	log.Printf("Server starting on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}

func uploadHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		log.Printf("Received upload request from %s", r.RemoteAddr)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()

		file, header, err := r.FormFile("file")
		if err != nil {
			log.Printf("Error retrieving file from request: %v", err)
			http.Error(w, "Invalid file upload", http.StatusBadRequest)
			return
		}
		defer file.Close()
		log.Printf("File received: %s, size: %d bytes", header.Filename, header.Size)

		log.Printf("Starting PDF content extraction for %s", header.Filename)
		content, err := pdf.ExtractContent(file)
		if err != nil {
			log.Printf("PDF extraction failed for %s: %v", header.Filename, err)
			http.Error(w, "PDF processing failed", http.StatusInternalServerError)
			return
		}
		inputWordCount := len(strings.Fields(content))
		log.Printf("PDF content extracted successfully, content length: %d words", inputWordCount)

		log.Printf("Chunking text with chunk size %d words", cfg.ChunkSize)
		chunks, err := chunker.ChunkText(content, cfg.ChunkSize)
		if err != nil {
			log.Printf("Text chunking failed: %v", err)
			http.Error(w, "Text chunking failed", http.StatusInternalServerError)
			return
		}
		log.Printf("Text successfully chunked into %d parts", len(chunks))

		log.Printf("Starting processing of %d chunks with max concurrency %d", len(chunks), cfg.MaxConcurrent)
		results := workers.ProcessChunks(ctx, chunks, cfg)
		if len(results) == 0 {
			log.Printf("Processing failed: no results returned")
			http.Error(w, "Processing failed", http.StatusInternalServerError)
			return
		}
		log.Printf("Successfully processed %d/%d chunks", len(results), len(chunks))

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-processed.txt", header.Filename))

		combinedResult := combineResults(results)
		outputWordCount := len(strings.Fields(combinedResult))
		reductionPercent := 100.0
		if inputWordCount > 0 {
			reductionPercent = 100.0 - (float64(outputWordCount) / float64(inputWordCount) * 100.0)
		}

		log.Printf("Sending response, combined result size: %d words (reduced from %d words, %.1f%% reduction)",
			outputWordCount, inputWordCount, reductionPercent)
		io.WriteString(w, combinedResult)

		log.Printf("Request completed in %v", time.Since(startTime))
	}
}

func combineResults(results []string) string {
	log.Printf("Combining %d result chunks", len(results))
	var final strings.Builder
	totalWords := 0

	for i, res := range results {
		wordCount := len(strings.Fields(res))
		totalWords += wordCount
		log.Printf("Chunk %d: %d words", i+1, wordCount)
		final.WriteString(res)
		final.WriteString("\n\n")
	}

	log.Printf("Combined %d chunks into %d words total", len(results), totalWords)
	return final.String()
}
