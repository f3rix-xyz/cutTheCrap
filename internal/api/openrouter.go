package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func ProcessText(ctx context.Context, text, apiKey string, targetWordCount int) (string, error) {
	startTime := time.Now()
	inputWordCount := len(strings.Fields(text))
	log.Printf("Processing text chunk of %d words (target: %d words)", inputWordCount, targetWordCount)

	prompt := fmt.Sprintf(`Condense this text to approximately %d words while:
- Preserving all key plot points and essential information
- Removing redundant descriptions and unnecessary elaborations
- Using extremely simple English with basic vocabulary (like for a 10-year-old)
- Using short, simple sentences without complex structures
- Avoiding any advanced vocabulary, idioms, or complicated expressions
- Maintaining the original narrative flow and storytelling style
- Keeping the text engaging and interesting

Important: Return ONLY the condensed text without any introductions, explanations, or summaries. Do not include phrases like "Here's the condensed version" or "In summary". Just provide the rewritten text directly.`, targetWordCount)

	payload := map[string]interface{}{
		"model": "deepseek/deepseek-chat:free",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": fmt.Sprintf("%s\n\n%s", text, prompt),
			},
		},
	}

	body, _ := json.Marshal(payload)
	log.Printf("Preparing API request to OpenRouter with model deepseek/deepseek-chat:free")
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	log.Printf("Sending request to OpenRouter API")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("API request failed: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API request returned non-OK status: %s", resp.Status)
		return "", fmt.Errorf("API request failed: %s", resp.Status)
	}
	log.Printf("Received response from OpenRouter API with status %s", resp.Status)

	var response OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Failed to decode API response: %v", err)
		return "", err
	}

	if len(response.Choices) == 0 {
		log.Printf("API response contained no choices")
		return "", fmt.Errorf("no content in response")
	}

	result := response.Choices[0].Message.Content
	outputWordCount := len(strings.Fields(result))
	reductionPercent := 100.0
	if inputWordCount > 0 {
		reductionPercent = 100.0 - (float64(outputWordCount) / float64(inputWordCount) * 100.0)
	}

	log.Printf("Successfully processed text in %v, result length: %d words (reduced from %d words, %.1f%% reduction)",
		time.Since(startTime), outputWordCount, inputWordCount, reductionPercent)
	return result, nil
}
