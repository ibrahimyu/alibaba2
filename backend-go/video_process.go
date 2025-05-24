package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// VideoFormData represents the structure of input data for video generation
type VideoInputData struct {
	RestoName    string     `json:"resto_name"`
	RestoAddress string     `json:"resto_address"`
	OpeningScene Scene      `json:"opening_scene"`
	ClosingScene Scene      `json:"closing_scene"`
	Music        Music      `json:"music"`
	Menu         []MenuItem `json:"menu"`
}

// VideoGenerationResult represents the result of the video generation process
type VideoGenerationResult struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	VideoPath    string `json:"video_path,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// GenerateVideo generates a restaurant promotional video based on the provided input data
func GenerateVideo(inputFile string, outputDir string, progressCallback func(stage string, percent int, message string)) (*VideoGenerationResult, error) {
	// Initialize result
	result := &VideoGenerationResult{
		Success: false,
	}

	// Call progress callback if provided
	if progressCallback != nil {
		progressCallback("init", 5, "Initializing video generation")
	}

	// Create output directory if it doesn't exist
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}

	tempDir := filepath.Join(outputDir, "temp")
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		os.MkdirAll(tempDir, 0755)
	}

	// Read and parse input JSON
	inputData, err := readInputData(inputFile)
	if err != nil {
		result.Message = "Failed to read input data"
		result.ErrorMessage = err.Error()
		return result, err
	}

	// Create API client using environment variable for API key
	apiKey := os.Getenv("ALIBABA_API_KEY")
	if apiKey == "" {
		err := fmt.Errorf("ALIBABA_API_KEY environment variable not set")
		result.Message = "Missing API key"
		result.ErrorMessage = err.Error()
		return result, err
	}
	alibabaAPI := NewAlibabaAPI(apiKey)

	// Update progress if callback provided
	if progressCallback != nil {
		progressCallback("segments", 20, "Generating video segments")
	}

	// Generate video segments
	videoSegments, err := generateVideoSegments(alibabaAPI, inputData, tempDir, progressCallback)
	if err != nil {
		result.Message = "Failed to generate video segments"
		result.ErrorMessage = err.Error()
		return result, err
	}

	// Generate background music if enabled
	var musicPath string
	if inputData.Music.Enabled {
		if progressCallback != nil {
			progressCallback("music", 50, "Generating background music")
		}

		musicPath, err = generateBackgroundMusic(inputData, outputDir)
		if err != nil {
			log.Printf("Warning: Failed to generate background music: %v", err)
			// Continue without music if generation fails
		}
	}

	// Combine videos and add music
	if progressCallback != nil {
		progressCallback("combining", 80, "Combining videos and adding music")
	}

	finalVideo, err := combineVideosWithMusic(videoSegments, musicPath, outputDir)
	if err != nil {
		result.Message = "Failed to combine videos"
		result.ErrorMessage = err.Error()
		return result, err
	}

	// Mark as completed
	if progressCallback != nil {
		progressCallback("finalizing", 95, "Finalizing video")
	}

	result.Success = true
	result.Message = "Video generation completed successfully"
	result.VideoPath = finalVideo

	return result, nil
}

// readInputData reads and parses the input JSON file
func readInputData(inputFile string) (*VideoInputData, error) {
	// Read file
	fileData, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("error reading input file: %w", err)
	}

	// Parse JSON
	var inputData VideoInputData
	if err := json.Unmarshal(fileData, &inputData); err != nil {
		return nil, fmt.Errorf("error parsing input JSON: %w", err)
	}

	return &inputData, nil
}

// generateVideoSegments generates all video segments (opening, menu items, closing)
func generateVideoSegments(api *AlibabaAPI, inputData *VideoInputData, tempDir string, progressCallback func(stage string, percent int, message string)) ([]string, error) {
	videoSegmentPaths := make([]string, 2+len(inputData.Menu)) // opening + menu items + closing
	var wg sync.WaitGroup
	var mutex sync.Mutex
	errorsChan := make(chan error, 2+len(inputData.Menu))

	// 1. Generate opening segment
	wg.Add(1)
	go func() {
		defer wg.Done()
		openingVideoPath, err := generatePromptVideo(api, inputData.OpeningScene.Prompt, inputData.OpeningScene.ImageURL, "opening", tempDir)
		if err != nil {
			errorsChan <- fmt.Errorf("failed to generate opening video: %w", err)
			return
		}

		mutex.Lock()
		videoSegmentPaths[0] = openingVideoPath
		mutex.Unlock()

		if progressCallback != nil {
			progressCallback("segments", 30, "Opening scene generated")
		}
	}()

	// 2. Generate menu item segments
	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	for i, menuItem := range inputData.Menu {
		wg.Add(1)
		i, menuItem := i, menuItem // Create local copies for the closure

		go func() {
			defer wg.Done()

			// Generate narration
			var menuNarration string
			var err error

			if openaiApiKey != "" {
				menuNarration, err = generateFoodVideoNarration(openaiApiKey, NarrationRequest{
					FoodName:        menuItem.Name,
					FoodDescription: menuItem.Description,
					NarrationLength: "short",
					ToneOfVoice:     "enthusiastic",
				})

				if err != nil {
					log.Printf("Warning: Failed to generate narration for menu item %s: %v", menuItem.Name, err)
					menuNarration = fmt.Sprintf("Try our delicious %s. %s", menuItem.Name, menuItem.Description)
				}
			} else {
				menuNarration = fmt.Sprintf("Try our delicious %s. %s", menuItem.Name, menuItem.Description)
			}

			// Generate video
			menuVideoPath, err := generatePromptVideo(api, menuNarration, menuItem.PhotoURL, fmt.Sprintf("menu_%d", i), tempDir)
			if err != nil {
				errorsChan <- fmt.Errorf("failed to generate menu video for item %d: %w", i, err)
				return
			}

			mutex.Lock()
			videoSegmentPaths[i+1] = menuVideoPath // +1 because index 0 is for opening
			mutex.Unlock()

			if progressCallback != nil {
				progress := 30 + (40 * (i + 1) / len(inputData.Menu))
				progressCallback("segments", progress, fmt.Sprintf("Menu item %d/%d generated", i+1, len(inputData.Menu)))
			}
		}()
	}

	// 3. Generate closing segment
	wg.Add(1)
	go func() {
		defer wg.Done()
		closingVideoPath, err := generatePromptVideo(api, inputData.ClosingScene.Prompt, inputData.ClosingScene.ImageURL, "closing", tempDir)
		if err != nil {
			errorsChan <- fmt.Errorf("failed to generate closing video: %w", err)
			return
		}

		mutex.Lock()
		videoSegmentPaths[len(videoSegmentPaths)-1] = closingVideoPath // Last index
		mutex.Unlock()

		if progressCallback != nil {
			progressCallback("segments", 70, "Closing scene generated")
		}
	}()

	// Wait for all video generation to complete
	wg.Wait()
	close(errorsChan)

	// Check for errors
	for err := range errorsChan {
		return nil, err
	}

	// Filter out any empty paths (should not happen if all went well)
	filteredPaths := make([]string, 0, len(videoSegmentPaths))
	for _, path := range videoSegmentPaths {
		if path != "" {
			filteredPaths = append(filteredPaths, path)
		}
	}

	return filteredPaths, nil
}

// generatePromptVideo generates a single video segment with prompt and image
func generatePromptVideo(api *AlibabaAPI, prompt string, imageURL string, segmentName string, outputDir string) (string, error) {
	log.Printf("Generating video segment: %s", segmentName)

	// Generate video using API
	response, err := api.GenerateVideoFromImage(GenerateVideoFromImageParams{
		Prompt:       prompt,
		ImageURL:     imageURL,
		Resolution:   "720P",
		PromptExtend: true,
	})
	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	// Get task ID and check progress
	taskID := response.Output.TaskID
	log.Printf("Video task started for segment %s, task ID: %s", segmentName, taskID)

	// Poll for completion
	result, err := api.PollTaskCompletion(taskID, PollTaskCompletionParams{
		MaxAttempts: 30,
		IntervalMs:  30000, // 30 seconds
	})
	if err != nil {
		return "", fmt.Errorf("polling failed: %w", err)
	}

	if result.Output.TaskStatus != "SUCCEEDED" {
		return "", fmt.Errorf("video generation failed: %s", result.Output.Message)
	}

	// Download the video file
	videoURL := result.Output.VideoURL
	videoPath := filepath.Join(outputDir, fmt.Sprintf("%s.mp4", segmentName))

	if err := DownloadFile(videoURL, videoPath); err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}

	return videoPath, nil
}

// generateBackgroundMusic generates background music based on restaurant theme using YuE AI model
func generateBackgroundMusic(inputData *VideoInputData, outputDir string) (string, error) {
	// Extract genre tags
	genreTags := inputData.Music.Genres
	if genreTags == "" {
		genreTags = "ambient instrumental lounge"
	}

	// Add BPM tag if specified
	if inputData.Music.BPM > 0 {
		genreTags += fmt.Sprintf(" %dbpm", inputData.Music.BPM)
	}

	// Create lyrics content
	lyrics := inputData.Music.Lyrics
	if lyrics == "" {
		lyrics = fmt.Sprintf("%s - %s\n%s",
			inputData.RestoName,
			inputData.RestoAddress,
			inputData.OpeningScene.Prompt)
	}

	// Create a directory for music output
	musicDir := filepath.Join(outputDir, "music")
	if _, err := os.Stat(musicDir); os.IsNotExist(err) {
		os.MkdirAll(musicDir, 0755)
	}

	log.Printf("Generating music with YuE using genres: %s", genreTags)

	// Call the YuE music generation function
	musicFilePath, err := GenerateYuEMusic(genreTags, lyrics, musicDir)
	if err != nil {
		log.Printf("Error generating music with YuE: %v", err)
		return "", fmt.Errorf("failed to generate music with YuE: %w", err)
	}

	log.Printf("Generated music file: %s", musicFilePath)
	return musicFilePath, nil
}

// combineVideosWithMusic combines all video segments and adds background music
func combineVideosWithMusic(videoSegmentPaths []string, musicPath string, outputDir string) (string, error) {
	if len(videoSegmentPaths) == 0 {
		return "", fmt.Errorf("no video segments to combine")
	}

	// Create file list for FFmpeg
	fileListPath := filepath.Join(outputDir, "filelist.txt")
	fileListContent := ""
	for _, videoPath := range videoSegmentPaths {
		fileListContent += fmt.Sprintf("file '%s'\n", videoPath)
	}

	if err := os.WriteFile(fileListPath, []byte(fileListContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file list: %w", err)
	}

	// Combine videos
	combinedVideoPath := filepath.Join(outputDir, "combined_video.mp4")
	combineCmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", fileListPath, "-c", "copy", combinedVideoPath)

	output, err := combineCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to combine videos: %w, output: %s", err, string(output))
	}

	finalVideoPath := filepath.Join(outputDir, "final_video.mp4")

	// Add background music if provided
	if musicPath != "" && musicPath != "not-implemented" {
		musicCmd := exec.Command("ffmpeg", "-i", combinedVideoPath, "-i", musicPath,
			"-map", "0:v", "-map", "1:a", "-shortest", "-c:v", "copy",
			"-c:a", "aac", "-b:a", "192k", finalVideoPath)

		output, err := musicCmd.CombinedOutput()
		if err != nil {
			// If adding music fails, just use the combined video
			log.Printf("Warning: Failed to add background music: %v, output: %s", err, string(output))
			// Copy the combined video as the final video
			if err := copyFile(combinedVideoPath, finalVideoPath); err != nil {
				return "", fmt.Errorf("failed to copy combined video: %w", err)
			}
		}
	} else {
		// If no music, just rename the combined video
		if err := copyFile(combinedVideoPath, finalVideoPath); err != nil {
			return "", fmt.Errorf("failed to copy combined video: %w", err)
		}
	}

	return finalVideoPath, nil
}

// copyFile copies a file from src to dst
func copyFile(src string, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err = ioutil.WriteFile(dst, input, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}
