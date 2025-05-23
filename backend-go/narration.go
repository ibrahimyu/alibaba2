package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// NarrationRequest represents the parameters for generating food video narration
type NarrationRequest struct {
	FoodName        string
	FoodDescription string
	NarrationLength string // "short", "medium", or "long"
	ToneOfVoice     string // "enthusiastic", "professional", etc.
}

// generateFoodVideoNarration generates narration for food videos using OpenAI API
func generateFoodVideoNarration(apiKey string, req NarrationRequest) (string, error) {
	// Set default values if not provided
	if req.NarrationLength == "" {
		req.NarrationLength = "short"
	}
	if req.ToneOfVoice == "" {
		req.ToneOfVoice = "enthusiastic"
	}

	// Build the prompt
	prompt := fmt.Sprintf(`Create a %s, %s narration for a food video about "%s". 
The narration should highlight these details: %s. 
The narration should be engaging and suitable for a restaurant promotional video.
Keep it concise and focused on making the viewer hungry.`,
		req.NarrationLength, req.ToneOfVoice, req.FoodName, req.FoodDescription)

	// Create the request body
	requestBody := map[string]interface{}{
		"model": "gpt-4-turbo",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are an expert copywriter specializing in food videos. Create concise, mouthwatering narrations that highlight the best qualities of dishes.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":        150,
		"temperature":       0.7,
		"top_p":             1,
		"presence_penalty":  0,
		"frequency_penalty": 0,
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling OpenAI request: %w", err)
	}

	// Create HTTP request
	req2, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req2)
	if err != nil {
		return "", fmt.Errorf("error sending request to OpenAI: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return "", fmt.Errorf("OpenAI API request failed with status %d", resp.StatusCode)
		}
		return "", fmt.Errorf("OpenAI API request failed: %v", errorResponse)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error parsing OpenAI response: %w", err)
	}

	// Extract the generated narration
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format from OpenAI")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid choice format in OpenAI response")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid message format in OpenAI response")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid content format in OpenAI response")
	}

	// Clean up the narration - remove quotes if present
	narration := strings.TrimSpace(content)
	narration = strings.Trim(narration, "\"")

	return narration, nil
}
