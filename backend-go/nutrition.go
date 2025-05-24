package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	Role    string `json:"role"`
	Content string `json:"content"`
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

	// Create HTTP request
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-DashScope-SSE", "disable") // Disable server-sent events for simpler processing

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response DashscopeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	// Check for API-level errors
	if response.Code != "" && response.Code != "Success" {
		return nil, fmt.Errorf("API error: %s - %s", response.Code, response.Message)
	}

	// Ensure we have content
	if len(response.Output.Choices) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// Extract text content from response
	rawResponse := response.Output.Choices[0].Message.Content

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

	// Split the text by lines for processing
	lines := strings.Split(rawText, "\n")

	var currentFood *FoodItem
	var inTotalSection bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and headers
		if line == "" || strings.Contains(line, "**Nutritional Content") {
			continue
		}

		// Check if we're starting a new food item
		if match := strings.Index(line, "**"); match == 0 {
			// Save previous food if it exists
			if currentFood != nil {
				result.Foods = append(result.Foods, *currentFood)
			}

			// Start new food item
			nameParts := strings.SplitN(strings.Trim(line, "* "), "(", 2)
			name := strings.TrimSpace(nameParts[0])

			serving := ""
			if len(nameParts) > 1 {
				serving = strings.TrimSpace(strings.TrimSuffix(nameParts[1], "):"))
			}

			currentFood = &FoodItem{
				Name:    name,
				Serving: serving,
			}

			inTotalSection = strings.Contains(line, "**Total") || strings.Contains(line, "Total Estimated")
			continue
		}

		// Process nutrition lines
		if currentFood != nil && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				nutrient := strings.ToLower(strings.TrimSpace(parts[0]))
				value := strings.TrimSpace(parts[1])

				// Remove leading dash if present
				if strings.HasPrefix(nutrient, "-") {
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
