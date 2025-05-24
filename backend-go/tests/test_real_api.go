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

	"github.com/joho/godotenv"
)

// This is a test file for directly testing the Dashscope API with a real image
// Run with: go run tests/test_real_api.go

// importNeeded copies the same structures from nutrition.go
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

func main() {
	// Load .env file
	err := godotenv.Load("../.env")
	if err != nil {
		// Try loading from current directory as fallback
		err = godotenv.Load(".env")
		if err != nil {
			fmt.Println("Error loading .env file:", err)
			return
		}
	}

	// Use a test image URL - replace with a valid food image URL
	imageURL := "https://example.com/test_food_image.jpg"

	// Optional: Check for command line arguments for image URL
	if len(os.Args) > 1 {
		imageURL = os.Args[1]
	}

	fmt.Printf("Testing food nutrition analysis with image: %s\n", imageURL)

	// Call the API directly for testing
	result, err := testFoodAnalysis(imageURL)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	fmt.Println("\n===== FINAL EXTRACTED CONTENT =====")
	fmt.Println(result)
}

// testFoodAnalysis is a test function that directly calls the Dashscope API
func testFoodAnalysis(imageURL string) (string, error) {
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("DASHSCOPE_API_KEY not found in environment")
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
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	fmt.Printf("Request payload: %s\n", string(jsonData))

	// Create HTTP request
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
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
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check for HTTP-level errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Debug: Print out all headers
	fmt.Println("Response headers:")
	for name, values := range resp.Header {
		fmt.Printf("%s: %s\n", name, strings.Join(values, ", "))
	}

	// Process the streaming SSE response
	scanner := bufio.NewScanner(resp.Body)
	var accumulator strings.Builder
	var lineCount int

	fmt.Println("\n===== START OF STREAMING RESPONSE =====")
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()

		fmt.Printf("Line %d: %s\n", lineCount, line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle SSE prefix
		if !strings.HasPrefix(line, "data: ") {
			fmt.Println("Non-data line:", line)
			continue
		}

		// Remove "data: " prefix
		jsonData := strings.TrimPrefix(line, "data: ")

		// Check for "[DONE]" marker
		if jsonData == "[DONE]" {
			fmt.Println("[DONE] marker received, ending stream")
			break
		}

		// Parse the line as JSON (for debugging)
		var streamResponse DashscopeResponse
		if err := json.Unmarshal([]byte(jsonData), &streamResponse); err != nil {
			fmt.Printf("Warning: Error parsing JSON: %v\n", err)
			continue
		}

		// Extract the text if present
		if len(streamResponse.Output.Choices) > 0 &&
			len(streamResponse.Output.Choices[0].Message.Content) > 0 {

			contentMap := streamResponse.Output.Choices[0].Message.Content[0]
			text, ok := contentMap["text"]

			if ok {
				textValue, isString := text.(string)
				if isString {
					fmt.Printf("Extracted text: %q\n", textValue)

					// Check if it's thinking content or actual nutritional info
					if strings.Contains(textValue, "Thinking:") ||
						strings.Contains(textValue, "I'm analyzing") ||
						strings.Contains(textValue, "Let me look") {
						fmt.Println("-> Thinking content detected, skipping")
					} else {
						accumulator.WriteString(textValue)
					}
				}
			}
		}
	}
	fmt.Println("===== END OF STREAMING RESPONSE =====\n")

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading stream: %v", err)
	}

	return accumulator.String(), nil
}
