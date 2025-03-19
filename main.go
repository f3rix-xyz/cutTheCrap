// main.go
package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"pdf-processor/internal/chunker"
	"pdf-processor/internal/config"
	"pdf-processor/internal/workers"
	"strconv"
	"strings"
	"time"
)

// enableCors adds the necessary CORS headers to allow cross-origin requests
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*") // Allow any origin for development
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	log.Println("Starting PDF processor service")
	cfg := config.Load()
	log.Printf("Configuration loaded: Port=%s, MaxConcurrent=%d, ChunkSize=%d", cfg.Port, cfg.MaxConcurrent, cfg.ChunkSize)

	// Handle both OPTIONS preflight and actual processing
	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		// Always enable CORS headers
		enableCors(&w)

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// For other methods, proceed with normal processing
		uploadHandler(cfg)(w, r)
	})

	log.Printf("Server starting on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}

func uploadHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		log.Printf("Received upload request from %s", r.RemoteAddr)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()

		text := r.FormValue("text")
		if text == "" {
			log.Printf("Error: text field is missing in request")
			http.Error(w, "Text field is missing", http.StatusBadRequest)
			return
		}

		ratioStr := r.FormValue("ratio")
		if ratioStr == "" {
			log.Printf("Error: ratio field is missing in request")
			http.Error(w, "Ratio field is missing", http.StatusBadRequest)
			return
		}

		ratio, err := strconv.ParseFloat(ratioStr, 64)
		if err != nil || ratio <= 0 || ratio > 1 {
			log.Printf("Error: invalid ratio value: %v", err)
			http.Error(w, "Invalid ratio value", http.StatusBadRequest)
			return
		}

		inputWordCount := len(strings.Fields(text))
		log.Printf("Received text, content length: %d words", inputWordCount)

		log.Printf("Chunking text with chunk size %d words", cfg.ChunkSize)
		chunks, err := chunker.ChunkText(text, cfg.ChunkSize)
		if err != nil {
			log.Printf("Text chunking failed: %v", err)
			http.Error(w, "Text chunking failed", http.StatusInternalServerError)
			return
		}
		log.Printf("Text successfully chunked into %d parts", len(chunks))

		log.Printf("Starting processing of %d chunks with max concurrency %d", len(chunks), cfg.MaxConcurrent)
		results := workers.ProcessChunks(ctx, chunks, cfg, ratio)
		if len(results) == 0 {
			log.Printf("Processing failed: no results returned")
			http.Error(w, "Processing failed", http.StatusInternalServerError)
			return
		}
		log.Printf("Successfully processed %d/%d chunks", len(results), len(chunks))

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", "attachment; filename=processed.txt")

		combinedResult := combineResults(results)
		outputWordCount := len(strings.Fields(combinedResult))
		reductionPercent := 100.0
		if inputWordCount > 0 {
			reductionPercent = 100.0 - (float64(outputWordCount)/float64(inputWordCount))*100.0
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
