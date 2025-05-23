package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// AlibabaAPI handles communication with the Alibaba Cloud DashScope API
type AlibabaAPI struct {
	APIKey   string
	BaseURL  string
	TasksURL string
}

// VideoGenerationResponse represents the response from the initial video generation API call
type VideoGenerationResponse struct {
	Output struct {
		TaskStatus string `json:"task_status"`
		TaskID     string `json:"task_id"`
	} `json:"output"`
	RequestID string `json:"request_id"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
}

// TaskStatusResponse represents the response from checking task status
type TaskStatusResponse struct {
	RequestID string `json:"request_id"`
	Output    struct {
		TaskID        string `json:"task_id"`
		TaskStatus    string `json:"task_status"`
		SubmitTime    string `json:"submit_time,omitempty"`
		ScheduledTime string `json:"scheduled_time,omitempty"`
		EndTime       string `json:"end_time,omitempty"`
		VideoURL      string `json:"video_url,omitempty"`
		Code          string `json:"code,omitempty"`
		Message       string `json:"message,omitempty"`
	} `json:"output"`
	Usage *struct {
		VideoDuration int    `json:"video_duration"`
		VideoRatio    string `json:"video_ratio"`
		VideoCount    int    `json:"video_count"`
	} `json:"usage,omitempty"`
}

// NewAlibabaAPI creates a new instance of the AlibabaAPI client
func NewAlibabaAPI(apiKey string) *AlibabaAPI {
	return &AlibabaAPI{
		APIKey:   apiKey,
		BaseURL:  "https://dashscope-intl.aliyuncs.com/api/v1/services/aigc/video-generation/video-synthesis",
		TasksURL: "https://dashscope-intl.aliyuncs.com/api/v1/tasks",
	}
}

// GenerateVideoFromImageParams represents the parameters for generating a video from an image
type GenerateVideoFromImageParams struct {
	Prompt       string
	ImageURL     string
	Resolution   string
	PromptExtend bool
}

// GenerateVideoFromImage generates a video from an image using Alibaba Cloud's DashScope API
func (api *AlibabaAPI) GenerateVideoFromImage(params GenerateVideoFromImageParams) (*VideoGenerationResponse, error) {
	if params.Resolution == "" {
		params.Resolution = "720P"
	}

	// Create request body
	reqBody := map[string]interface{}{
		"model": "wan2.1-i2v-turbo",
		"input": map[string]interface{}{
			"prompt":  params.Prompt,
			"img_url": params.ImageURL,
		},
		"parameters": map[string]interface{}{
			"resolution":    params.Resolution,
			"prompt_extend": params.PromptExtend,
		},
	}

	// Convert request body to JSON
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", api.BaseURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DashScope-Async", "enable")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.APIKey))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var result VideoGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
}

// GenerateVideoFromMultipleImagesParams represents the parameters for generating a video from multiple reference images
type GenerateVideoFromMultipleImagesParams struct {
	Prompt        string
	RefImagesURLs []string
	ObjOrBg       []string // "obj" or "bg"
	Size          string
}

// GenerateVideoFromMultipleImages generates a video from multiple reference images
func (api *AlibabaAPI) GenerateVideoFromMultipleImages(params GenerateVideoFromMultipleImagesParams) (*VideoGenerationResponse, error) {
	if params.Size == "" {
		params.Size = "1280*720"
	}

	if len(params.ObjOrBg) == 0 {
		params.ObjOrBg = []string{"obj", "bg"}
	}

	// Create request body
	reqBody := map[string]interface{}{
		"model": "wan2.1-vace-plus",
		"input": map[string]interface{}{
			"function":       "image_reference",
			"prompt":         params.Prompt,
			"ref_images_url": params.RefImagesURLs,
		},
		"parameters": map[string]interface{}{
			"obj_or_bg": params.ObjOrBg,
			"size":      params.Size,
		},
	}

	// Convert request body to JSON
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", api.BaseURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DashScope-Async", "enable")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.APIKey))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var result VideoGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
}

// GenerateVideoRepaintingParams represents the parameters for generating a video by repainting an existing video
type GenerateVideoRepaintingParams struct {
	Prompt           string
	VideoURL         string
	ControlCondition string
}

// GenerateVideoRepainting generates a video by repainting an existing video
func (api *AlibabaAPI) GenerateVideoRepainting(params GenerateVideoRepaintingParams) (*VideoGenerationResponse, error) {
	if params.ControlCondition == "" {
		params.ControlCondition = "depth"
	}

	// Create request body
	reqBody := map[string]interface{}{
		"model": "wan2.1-vace-plus",
		"input": map[string]interface{}{
			"function":  "video_repainting",
			"prompt":    params.Prompt,
			"video_url": params.VideoURL,
		},
		"parameters": map[string]interface{}{
			"control_condition": params.ControlCondition,
		},
	}

	// Convert request body to JSON
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", api.BaseURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DashScope-Async", "enable")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.APIKey))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var result VideoGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
}

// CheckTaskStatus checks the status of a video generation task
func (api *AlibabaAPI) CheckTaskStatus(taskID string) (*TaskStatusResponse, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", api.TasksURL, taskID), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.APIKey))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var result TaskStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
}

// PollTaskCompletionParams represents the parameters for polling a task until completion
type PollTaskCompletionParams struct {
	MaxAttempts int
	IntervalMs  int
}

// PollTaskCompletion polls for task completion with configurable intervals
func (api *AlibabaAPI) PollTaskCompletion(taskID string, params PollTaskCompletionParams) (*TaskStatusResponse, error) {
	if params.MaxAttempts == 0 {
		params.MaxAttempts = 30
	}
	if params.IntervalMs == 0 {
		params.IntervalMs = 30000 // 30 seconds default
	}

	for attempts := 0; attempts < params.MaxAttempts; attempts++ {
		response, err := api.CheckTaskStatus(taskID)
		if err != nil {
			return nil, err
		}

		if response.Output.TaskStatus == "SUCCEEDED" || response.Output.TaskStatus == "FAILED" {
			return response, nil
		}

		// Wait for the specified interval
		time.Sleep(time.Duration(params.IntervalMs) * time.Millisecond)
	}

	return nil, errors.New("task polling timed out after maximum attempts")
}

// DownloadFile downloads a file from a URL and saves it to the specified path
func DownloadFile(url string, destPath string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(destPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}

	// Create a new HTTP client
	client := &http.Client{}

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error downloading file: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create the output file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer out.Close()

	// Copy the response body to the output file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	log.Printf("Downloaded file to %s", destPath)
	return nil
}
