package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// DashscopeRequest represents the structure for the Dashscope API request
type DashscopeRequest struct {
	Model      string               `json:"model"`
	Input      DashscopeInput       `json:"input"`
	Parameters *DashscopeParameters `json:"parameters,omitempty"`
}

// DashscopeInput represents the input part of the Dashscope request
type DashscopeInput struct {
	Messages []DashscopeMessage `json:"messages"`
}

// DashscopeMessage represents a message in the Dashscope conversation
type DashscopeMessage struct {
	Role    string                    `json:"role"`
	Content []DashscopeMessageContent `json:"content"`
}

// DashscopeMessageContent represents content in a message, either text or image
type DashscopeMessageContent struct {
	Text  string `json:"text,omitempty"`
	Image string `json:"image,omitempty"`
}

// DashscopeParameters represents optional parameters for the request
type DashscopeParameters struct {
	TopP        *float64 `json:"top_p,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
}

// DashscopeResponse represents the response from the Dashscope API
type DashscopeResponse struct {
	RequestID string          `json:"request_id"`
	Output    DashscopeOutput `json:"output"`
	Usage     DashscopeUsage  `json:"usage"`
	Code      string          `json:"code"`
	Message   string          `json:"message"`
}

// DashscopeOutput represents the output part of the response
type DashscopeOutput struct {
	FinishReason string            `json:"finish_reason"`
	Choices      []DashscopeChoice `json:"choices"`
}

// DashscopeChoice represents a choice in the Dashscope response
type DashscopeChoice struct {
	Message DashscopeChoiceMessage `json:"message"`
}

// DashscopeChoiceMessage represents the message in a choice
type DashscopeChoiceMessage struct {
	Role    string                   `json:"role"`
	Content []map[string]interface{} `json:"content"`
}

// DashscopeUsage represents the token usage information
type DashscopeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// FoodAnalysisResult represents the processed analysis result
type FoodAnalysisResult struct {
	Foods          []FoodItem       `json:"foods"`
	TotalNutrition NutritionSummary `json:"total_nutrition"`
	RawResponse    string           `json:"raw_response"`
}

// FoodItem represents a single food item and its nutritional content
type FoodItem struct {
	Name     string `json:"name"`
	Serving  string `json:"serving"`
	Calories string `json:"calories"`
	Fat      string `json:"fat"`
	Protein  string `json:"protein"`
	Carbs    string `json:"carbs"`
	Fiber    string `json:"fiber,omitempty"`
	Sodium   string `json:"sodium,omitempty"`
}

// NutritionSummary represents total nutritional content
type NutritionSummary struct {
	Calories string `json:"calories"`
	Fat      string `json:"fat"`
	Protein  string `json:"protein"`
	Carbs    string `json:"carbs"`
	Fiber    string `json:"fiber,omitempty"`
	Sodium   string `json:"sodium,omitempty"`
}

// AnalyzeFoodImage performs nutrition analysis on a food image
func AnalyzeFoodImage(imageURL string) (*FoodAnalysisResult, error) {
	// If the URL is relative, we need to convert to absolute URL
	if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		// Get the server's host
		host := os.Getenv("SERVER_HOST")
		if host == "" {
			host = "http://localhost:3000"
		}
		imageURL = host + imageURL
	}

	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		// If not found in env, use a default key (for development only)
		apiKey = "sk-xxx"
	}

	baseURL := "https://dashscope-intl.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"

	// Set temperature and top_p params
	temperature := 0.3
	topP := 0.8

	// Create Dashscope API request
	request := DashscopeRequest{
		Model: "qvq-max",
		Input: DashscopeInput{
			Messages: []DashscopeMessage{
				{
					Role: "user",
					Content: []DashscopeMessageContent{
						{
							Image: imageURL,
						},
						{
							Text: "make me what is the nutritional content with accurate numbers in this picture and what foods are in it and output just nutrition only",
						},
					},
				},
			},
		},
		Parameters: &DashscopeParameters{
			Temperature: &temperature,
			TopP:        &topP,
		},
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Debug log - enable if needed
	// fmt.Printf("Request payload: %s\n", string(jsonData))

	// Create HTTP request
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-DashScope-SSE", "enable")   // Enable server-sent events for streaming
	req.Header.Set("Accept", "text/event-stream") // Explicitly request SSE format

	// Set a longer timeout for streaming responses
	client := &http.Client{
		Timeout: 120 * time.Second, // 2 minute timeout
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check for HTTP-level errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Verify we got the expected content type for streaming
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		// Not a stream response, try to parse it as a regular JSON response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading non-streaming response: %v", err)
		}

		// Try to parse as normal JSON response
		var directResponse DashscopeResponse
		if err := json.Unmarshal(body, &directResponse); err == nil {
			// Successfully parsed as direct response
			if directResponse.Code != "" && directResponse.Code != "Success" && directResponse.Code != "0" {
				return nil, fmt.Errorf("API error: %s - %s", directResponse.Code, directResponse.Message)
			}

			if len(directResponse.Output.Choices) > 0 {
				// Extract the text from the content array
				var rawResponse string
				for _, contentItem := range directResponse.Output.Choices[0].Message.Content {
					if text, ok := contentItem["text"]; ok {
						if textStr, ok := text.(string); ok {
							rawResponse += textStr
						}
					}
				}

				result := processFoodAnalysis(rawResponse)
				result.RawResponse = rawResponse
				return result, nil
			}
		}

		return nil, fmt.Errorf("unexpected response format: %s", contentType)
	}

	// Process the streaming SSE response
	scanner := bufio.NewScanner(resp.Body)
	var finalResponse *DashscopeResponse
	var accumulator strings.Builder
	var isThinking bool

	for scanner.Scan() {
		line := scanner.Text()
		// Debug log only if needed
		// fmt.Printf("Received line: %s\n", line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle SSE prefix
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Remove "data: " prefix
		jsonData := strings.TrimPrefix(line, "data: ")

		// Check for "[DONE]" marker
		if jsonData == "[DONE]" {
			break
		}

		// Parse the line as JSON
		var streamResponse DashscopeResponse
		if err := json.Unmarshal([]byte(jsonData), &streamResponse); err != nil {
			fmt.Printf("Warning: Error parsing JSON: %v\n", err)
			continue // Skip parsing errors, just try the next line
		}

		// Check for API-level errors
		if streamResponse.Code != "" && streamResponse.Code != "Success" && streamResponse.Code != "0" {
			return nil, fmt.Errorf("API error: %s - %s", streamResponse.Code, streamResponse.Message)
		}

		// Skip empty responses
		if len(streamResponse.Output.Choices) == 0 {
			continue
		}

		// Track the latest response object for metadata
		finalResponse = &streamResponse

		// Extract the text chunks from the message content
		// The structure is: choices[0].message.content[0].text
		if len(streamResponse.Output.Choices[0].Message.Content) > 0 {
			// Access the text field safely to avoid potential panics
			contentMap := streamResponse.Output.Choices[0].Message.Content[0]
			text, ok := contentMap["text"]
			if !ok || text == nil {
				// If text is not found or nil, skip this chunk
				continue
			}

			// Extract the text value as a string
			textValue, ok := text.(string)
			if !ok {
				// If text is not a string, skip this chunk
				continue
			}

			// Check if we're in a thinking block
			if strings.Contains(textValue, "Thinking:") ||
				strings.Contains(textValue, "I'm analyzing") ||
				strings.Contains(textValue, "Let me look") {
				isThinking = true
				// Don't accumulate thinking content
				continue
			}

			// If we have nutritional content, or if we're not in a thinking block,
			// accumulate the content
			if !isThinking ||
				strings.Contains(textValue, "**Nutritional Content") ||
				strings.Contains(textValue, "Calories:") ||
				strings.Contains(textValue, "Protein:") {
				// Exit thinking mode if we encounter nutritional content
				if isThinking {
					isThinking = false
				}
				// Accumulate the text
				accumulator.WriteString(textValue)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading stream: %v", err)
	}

	// Use the accumulated content as the final response
	accumulatedContent := accumulator.String()

	// If we didn't get any content, that's an error
	if accumulatedContent == "" || finalResponse == nil {
		return nil, fmt.Errorf("no valid nutrition content found in response")
	}

	// Use the accumulated content
	rawResponse := accumulatedContent

	// Process the raw response into structured data
	result := processFoodAnalysis(rawResponse)
	result.RawResponse = rawResponse

	return result, nil
}

// processFoodAnalysis extracts structured nutrition information from raw text
func processFoodAnalysis(rawText string) *FoodAnalysisResult {
	result := &FoodAnalysisResult{
		Foods: make([]FoodItem, 0),
	}

	// Clean up the text - remove markdown artifacts that might interfere with parsing
	cleanText := strings.ReplaceAll(rawText, "\\n", "\n")
	cleanText = strings.ReplaceAll(cleanText, "\\*", "*")
	cleanText = strings.ReplaceAll(cleanText, "\\-", "-")

	// Split the text by lines for processing
	lines := strings.Split(cleanText, "\n")

	var currentFood *FoodItem
	var inTotalSection bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip header lines but note when we're entering the nutritional section
		if strings.Contains(line, "**Nutritional Content") ||
			strings.Contains(line, "*Nutritional Content") ||
			strings.Contains(line, "Nutritional Content") {
			continue
		}

		// Check if we're starting a new food item - handle various markdown formats
		if (strings.HasPrefix(line, "**") ||
			strings.HasPrefix(line, "*") ||
			strings.HasPrefix(line, "- ") ||
			strings.HasPrefix(line, "1. ")) &&
			(strings.Contains(line, ":") ||
				strings.Contains(line, ")") ||
				strings.Contains(line, "g") ||
				strings.Contains(line, "serving")) {

			// Save previous food if it exists
			if currentFood != nil {
				result.Foods = append(result.Foods, *currentFood)
			}

			// Clean up the line, removing any markdown symbols
			cleanLine := strings.TrimPrefix(line, "- ")
			cleanLine = strings.TrimPrefix(cleanLine, "* ")
			cleanLine = strings.TrimPrefix(cleanLine, "** ")
			cleanLine = strings.TrimPrefix(cleanLine, "1. ")
			cleanLine = strings.TrimPrefix(cleanLine, "**")
			cleanLine = strings.TrimSuffix(cleanLine, "**")

			// Start new food item
			nameParts := strings.SplitN(cleanLine, "(", 2)
			name := strings.TrimSpace(nameParts[0])

			// Handle any colon in the name
			if strings.Contains(name, ":") {
				name = strings.TrimSpace(strings.Split(name, ":")[0])
			}

			serving := ""
			if len(nameParts) > 1 {
				serving = strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(nameParts[1], "):"), ")"))
			}

			currentFood = &FoodItem{
				Name:    name,
				Serving: serving,
			}

			inTotalSection = strings.Contains(line, "Total") ||
				strings.Contains(line, "TOTAL") ||
				strings.HasPrefix(cleanLine, "Total")
			continue
		}

		// Process nutrition lines
		if currentFood != nil && strings.Contains(line, ":") {
			// Clean up any markdown artifacts
			cleanLine := strings.TrimPrefix(line, "- ")
			cleanLine = strings.TrimPrefix(cleanLine, "* ")
			cleanLine = strings.TrimPrefix(cleanLine, "** ")
			cleanLine = strings.TrimPrefix(cleanLine, "**")

			parts := strings.SplitN(cleanLine, ":", 2)
			if len(parts) == 2 {
				nutrient := strings.ToLower(strings.TrimSpace(parts[0]))
				value := strings.TrimSpace(parts[1])

				// Remove leading dash or bullet if present
				if strings.HasPrefix(nutrient, "-") {
					nutrient = strings.TrimSpace(nutrient[1:])
				}
				if strings.HasPrefix(nutrient, "•") {
					nutrient = strings.TrimSpace(nutrient[1:])
				}

				switch {
				case strings.Contains(nutrient, "calor"):
					if inTotalSection {
						result.TotalNutrition.Calories = value
					} else {
						currentFood.Calories = value
					}
				case strings.Contains(nutrient, "fat"):
					if inTotalSection {
						result.TotalNutrition.Fat = value
					} else {
						currentFood.Fat = value
					}
				case strings.Contains(nutrient, "protein"):
					if inTotalSection {
						result.TotalNutrition.Protein = value
					} else {
						currentFood.Protein = value
					}
				case strings.Contains(nutrient, "carb"):
					if inTotalSection {
						result.TotalNutrition.Carbs = value
					} else {
						currentFood.Carbs = value
					}
				case strings.Contains(nutrient, "fiber"):
					if inTotalSection {
						result.TotalNutrition.Fiber = value
					} else {
						currentFood.Fiber = value
					}
				case strings.Contains(nutrient, "sodium"):
					if inTotalSection {
						result.TotalNutrition.Sodium = value
					} else {
						currentFood.Sodium = value
					}
				}
			}
		}
	}

	// Add the last food if it exists and is not a total
	if currentFood != nil && !inTotalSection {
		result.Foods = append(result.Foods, *currentFood)
	}

	return result
}
