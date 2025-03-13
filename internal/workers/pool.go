package workers

import (
	"context"
	"pdf-processor/internal/api"
	"pdf-processor/internal/config"
	"sync"
)

func ProcessChunks(ctx context.Context, chunks []string, cfg *config.Config) []string {
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
		for i, chunk := range chunks {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(index int, text string) {
				defer func() {
					<-semaphore
					wg.Done()
				}()

				content, err := api.ProcessText(ctx, text, cfg.OpenRouterKey)
				if err == nil {
					resultChan <- struct {
						index   int
						content string
					}{index, content}
				}
			}(i, chunk)
		}
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		results[res.index] = res.content
	}
	return results
}
