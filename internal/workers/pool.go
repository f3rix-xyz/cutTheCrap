package workers

import (
	"context"
	"log"
	"pdf-processor/internal/api"
	"pdf-processor/internal/config"
	"strings"
	"sync"
	"time"
)

func ProcessChunks(ctx context.Context, chunks []string, cfg *config.Config, ratio float64) []string {
	startTime := time.Now()

	totalInputWords := 0
	for _, chunk := range chunks {
		totalInputWords += len(strings.Fields(chunk))
	}

	log.Printf("Starting to process %d chunks with max concurrency %d (total input: %d words)",
		len(chunks), cfg.MaxConcurrent, totalInputWords)

	var (
		wg         sync.WaitGroup
		results    = make([]string, len(chunks))
		semaphore  = make(chan struct{}, cfg.MaxConcurrent)
		resultChan = make(chan struct {
			index   int
			content string
		})
	)

	go func() {
		log.Printf("Worker goroutine started, will process %d chunks", len(chunks))
		for i, chunk := range chunks {
			wg.Add(1)
			semaphore <- struct{}{}
			chunkWords := len(strings.Fields(chunk))
			log.Printf("Dispatching worker for chunk %d/%d (size: %d words)", i+1, len(chunks), chunkWords)

			go func(index int, text string) {
				chunkStartTime := time.Now()
				defer func() {
					<-semaphore
					wg.Done()
					log.Printf("Worker for chunk %d completed in %v", index, time.Since(chunkStartTime))
				}()

				inputWords := len(strings.Fields(text))
				log.Printf("Processing chunk %d (%d words)", index, inputWords)

				targetWordCount := int(float64(cfg.ChunkSize) * ratio)
				if targetWordCount <= 0 {
					targetWordCount = 1
				}

				content, err := api.ProcessText(ctx, text, cfg.OpenRouterKey, targetWordCount)
				if err != nil {
					log.Printf("Error processing chunk %d: %v", index, err)
				} else {
					outputWords := len(strings.Fields(content))
					log.Printf("Successfully processed chunk %d, result: %d words", index, outputWords)
					resultChan <- struct {
						index   int
						content string
					}{index, content}
				}
			}(i, chunk)
		}
		log.Println("All workers dispatched, waiting for completion")
		wg.Wait()
		log.Println("All workers completed, closing result channel")
		close(resultChan)
	}()

	log.Println("Collecting results from workers")
	resultCount := 0
	for res := range resultChan {
		resultCount++
		resultWords := len(strings.Fields(res.content))
		log.Printf("Received result %d/%d for chunk %d (%d words)", resultCount, len(chunks), res.index, resultWords)
		results[res.index] = res.content
	}

	validResults := 0
	totalOutputWords := 0
	for _, r := range results {
		if r != "" {
			validResults++
			totalOutputWords += len(strings.Fields(r))
		}
	}

	reductionPercent := 100.0
	if totalInputWords > 0 {
		reductionPercent = 100.0 - (float64(totalOutputWords)/float64(totalInputWords))*100.0
	}

	log.Printf("Processing completed in %v, received %d valid results out of %d chunks",
		time.Since(startTime), validResults, len(chunks))
	log.Printf("Total input: %d words, total output: %d words (%.1f%% reduction)",
		totalInputWords, totalOutputWords, reductionPercent)

	return results
}
