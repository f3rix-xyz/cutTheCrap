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
	cfg := config.Load()
	http.HandleFunc("/process", uploadHandler(cfg))
	log.Printf("Server starting on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}

func uploadHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Invalid file upload", http.StatusBadRequest)
			return
		}
		defer file.Close()

		content, err := pdf.ExtractContent(file)
		if err != nil {
			http.Error(w, "PDF processing failed", http.StatusInternalServerError)
			return
		}

		chunks, err := chunker.ChunkText(content, cfg.ChunkSize)
		if err != nil {
			http.Error(w, "Text chunking failed", http.StatusInternalServerError)
			return
		}

		results := workers.ProcessChunks(ctx, chunks, cfg)
		if len(results) == 0 {
			http.Error(w, "Processing failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-processed.txt", header.Filename))
		io.WriteString(w, combineResults(results))
	}
}

func combineResults(results []string) string {
	var final strings.Builder
	for _, res := range results {
		final.WriteString(res)
		final.WriteString("\n\n")
	}
	return final.String()
}
