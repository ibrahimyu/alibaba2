package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
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

// VideoGenerationCheckpoint tracks which segments have been completed
type VideoGenerationCheckpoint struct {
	JobID             string            `json:"job_id"`
	CompletedSegments map[string]string `json:"completed_segments"` // segment_name -> file_path
	OpeningComplete   bool              `json:"opening_complete"`
	MenuItemsComplete map[int]bool      `json:"menu_items_complete"`
	ClosingComplete   bool              `json:"closing_complete"`
	MusicGenerated    bool              `json:"music_generated"`
	MusicPath         string            `json:"music_path,omitempty"`
	CheckpointTime    time.Time         `json:"checkpoint_time"`
}

// saveCheckpoint saves the current generation progress
func saveCheckpoint(checkpoint VideoGenerationCheckpoint, outputDir string) error {
	checkpointFile := filepath.Join(outputDir, "checkpoint.json")
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint data: %w", err)
	}

	return ioutil.WriteFile(checkpointFile, data, 0644)
}

// loadCheckpoint loads the previous generation progress if it exists
func loadCheckpoint(jobID string, outputDir string) (*VideoGenerationCheckpoint, error) {
	checkpointFile := filepath.Join(outputDir, "checkpoint.json")

	// If checkpoint doesn't exist, return a new one
	if _, err := os.Stat(checkpointFile); os.IsNotExist(err) {
		return &VideoGenerationCheckpoint{
			JobID:             jobID,
			CompletedSegments: make(map[string]string),
			MenuItemsComplete: make(map[int]bool),
			CheckpointTime:    time.Now(),
		}, nil
	}

	data, err := ioutil.ReadFile(checkpointFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %w", err)
	}

	var checkpoint VideoGenerationCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint data: %w", err)
	}

	return &checkpoint, nil
}

// GenerateVideo generates a restaurant promotional video based on the provided input data
func GenerateVideo(inputFile string, outputDir string, progressCallback func(stage string, percent int, message string)) (*VideoGenerationResult, error) {
	// Initialize result
	result := &VideoGenerationResult{
		Success: false,
	}

	// Extract jobID from outputDir for checkpoint tracking
	jobID := filepath.Base(outputDir)
	if strings.HasPrefix(jobID, "output_") {
		jobID = jobID[7:] // Remove "output_" prefix
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

	// Load checkpoint if exists
	checkpoint, err := loadCheckpoint(jobID, outputDir)
	if err != nil {
		log.Printf("Warning: Failed to load checkpoint: %v", err)
		// Create a new checkpoint
		checkpoint = &VideoGenerationCheckpoint{
			JobID:             jobID,
			CompletedSegments: make(map[string]string),
			MenuItemsComplete: make(map[int]bool),
			CheckpointTime:    time.Now(),
		}
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

	// Generate video segments with checkpoint support
	videoSegments, err := generateVideoSegments(alibabaAPI, inputData, tempDir, jobID, outputDir, progressCallback)
	if err != nil {
		result.Message = "Failed to generate video segments"
		result.ErrorMessage = err.Error()
		return result, err
	}

	// Generate background music if enabled and not already generated
	var musicPath string
	if inputData.Music.Enabled && !checkpoint.MusicGenerated {
		if progressCallback != nil {
			progressCallback("music", 50, "Generating background music")
		}

		musicPath, err = generateBackgroundMusic(inputData, outputDir)
		if err != nil {
			log.Printf("Warning: Failed to generate background music: %v", err)
			// Continue without music if generation fails
		} else {
			checkpoint.MusicGenerated = true
			checkpoint.MusicPath = musicPath
			saveCheckpoint(*checkpoint, outputDir)
		}
	} else if checkpoint.MusicGenerated {
		musicPath = checkpoint.MusicPath
		log.Printf("Using existing music: %s", musicPath)
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
func generateVideoSegments(api *AlibabaAPI, inputData *VideoInputData, tempDir string, jobID string, outputDir string, progressCallback func(stage string, percent int, message string)) ([]string, error) {
	// Load checkpoint if exists
	checkpoint, err := loadCheckpoint(jobID, outputDir)
	if err != nil {
		log.Printf("Warning: Failed to load checkpoint: %v", err)
		// Create a new checkpoint
		checkpoint = &VideoGenerationCheckpoint{
			JobID:             jobID,
			CompletedSegments: make(map[string]string),
			MenuItemsComplete: make(map[int]bool),
			CheckpointTime:    time.Now(),
		}
	}

	videoSegmentPaths := make([]string, 2+len(inputData.Menu)) // opening + menu items + closing
	var wg sync.WaitGroup
	var mutex sync.Mutex
	errorsChan := make(chan error, 2+len(inputData.Menu))

	// 1. Generate opening segment if not already completed
	if !checkpoint.OpeningComplete {
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
			checkpoint.OpeningComplete = true
			checkpoint.CompletedSegments["opening"] = openingVideoPath
			saveCheckpoint(*checkpoint, outputDir) // Save progress
			mutex.Unlock()

			if progressCallback != nil {
				progressCallback("segments", 30, "Opening scene generated")
			}
		}()
	} else {
		// Use existing video from checkpoint
		videoSegmentPaths[0] = checkpoint.CompletedSegments["opening"]
		log.Printf("Using existing opening segment: %s", videoSegmentPaths[0])
		if progressCallback != nil {
			progressCallback("segments", 30, "Using existing opening scene")
		}
	}

	// 2. Generate menu item segments
	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	for i, menuItem := range inputData.Menu {
		// Skip if this menu item is already completed
		if completed, exists := checkpoint.MenuItemsComplete[i]; exists && completed {
			segmentKey := fmt.Sprintf("menu_%d", i)
			videoSegmentPaths[i+1] = checkpoint.CompletedSegments[segmentKey]
			log.Printf("Using existing menu segment %d: %s", i, videoSegmentPaths[i+1])
			continue
		}

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
			segmentKey := fmt.Sprintf("menu_%d", i)
			menuVideoPath, err := generatePromptVideo(api, menuNarration, menuItem.PhotoURL, segmentKey, tempDir)
			if err != nil {
				errorsChan <- fmt.Errorf("failed to generate menu video for item %d: %w", i, err)
				return
			}

			mutex.Lock()
			videoSegmentPaths[i+1] = menuVideoPath // +1 because index 0 is for opening
			checkpoint.MenuItemsComplete[i] = true
			checkpoint.CompletedSegments[segmentKey] = menuVideoPath
			saveCheckpoint(*checkpoint, outputDir) // Save progress
			mutex.Unlock()

			if progressCallback != nil {
				progress := 30 + (40 * (i + 1) / len(inputData.Menu))
				progressCallback("segments", progress, fmt.Sprintf("Menu item %d/%d generated", i+1, len(inputData.Menu)))
			}
		}()
	}

	// 3. Generate closing segment if not already completed
	if !checkpoint.ClosingComplete {
		wg.Add(1)
		go func() {
			defer wg.Done()
			closingVideoPath, err := generatePromptVideo(api, inputData.ClosingScene.Prompt,
				inputData.ClosingScene.ImageURL, "closing", tempDir)
			if err != nil {
				errorsChan <- fmt.Errorf("failed to generate closing video: %w", err)
				return
			}

			mutex.Lock()
			videoSegmentPaths[len(videoSegmentPaths)-1] = closingVideoPath // Last index
			checkpoint.ClosingComplete = true
			checkpoint.CompletedSegments["closing"] = closingVideoPath
			saveCheckpoint(*checkpoint, outputDir) // Save progress
			mutex.Unlock()

			if progressCallback != nil {
				progressCallback("segments", 70, "Closing scene generated")
			}
		}()
	} else {
		// Use existing video from checkpoint
		videoSegmentPaths[len(videoSegmentPaths)-1] = checkpoint.CompletedSegments["closing"]
		log.Printf("Using existing closing segment: %s", videoSegmentPaths[len(videoSegmentPaths)-1])
		if progressCallback != nil {
			progressCallback("segments", 70, "Using existing closing scene")
		}
	}

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
		MaxAttempts: 300,
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

	// Handle special case with only one video
	if len(videoSegmentPaths) == 1 {
		combinedVideoPath := filepath.Join(outputDir, "combined_video.mp4")
		absVideoPath, err := filepath.Abs(videoSegmentPaths[0])
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for %s: %w", videoSegmentPaths[0], err)
		}
		absCombinedVideoPath, err := filepath.Abs(combinedVideoPath)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for combined video: %w", err)
		}

		// Just copy the single video
		copyCmd := exec.Command("ffmpeg", "-i", absVideoPath, "-c", "copy", absCombinedVideoPath)
		output, err := copyCmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to copy single video: %w, output: %s", err, string(output))
		}
	} else if len(videoSegmentPaths) > 1 {
		log.Printf("Combining %d video segments with fade transitions", len(videoSegmentPaths))

		// Create temp directory for transition processing
		tempDir := filepath.Join(outputDir, "temp_transitions")
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			os.MkdirAll(tempDir, 0755)
		}

		// Duration of transition in seconds
		transitionDuration := 1.0

		// Process videos with transitions
		combinedVideoPath := filepath.Join(outputDir, "combined_video.mp4")
		absCombinedVideoPath, err := filepath.Abs(combinedVideoPath)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for combined video: %w", err)
		}

		// Start with the first video
		firstVideoPath, err := filepath.Abs(videoSegmentPaths[0])
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for %s: %w", videoSegmentPaths[0], err)
		}

		// Copy the first video as starting point
		tempOutput := filepath.Join(tempDir, "temp_0.mp4")
		absFirstOutput, err := filepath.Abs(tempOutput)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for temp output: %w", err)
		}

		copyCmd := exec.Command("ffmpeg", "-i", firstVideoPath, "-c", "copy", absFirstOutput)
		output, err := copyCmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to copy first video: %w, output: %s", err, string(output))
		}

		// Process each successive video with a fade transition
		currentInput := absFirstOutput

		for i := 1; i < len(videoSegmentPaths); i++ {
			nextVideoPath, err := filepath.Abs(videoSegmentPaths[i])
			if err != nil {
				return "", fmt.Errorf("failed to get absolute path for %s: %w", videoSegmentPaths[i], err)
			}

			// Create next output path
			nextOutput := filepath.Join(tempDir, fmt.Sprintf("temp_%d.mp4", i))
			absNextOutput, err := filepath.Abs(nextOutput)
			if err != nil {
				return "", fmt.Errorf("failed to get absolute path for next output: %w", err)
			}

			// Get duration of current video
			duration := getVideoDuration(currentInput)
			offsetTime := duration - transitionDuration
			if offsetTime < 0 {
				offsetTime = 0
			}

			// Apply xfade filter for fade transition
			xfadeCmd := exec.Command("ffmpeg",
				"-i", currentInput,
				"-i", nextVideoPath,
				"-filter_complex", fmt.Sprintf("[0:v][1:v]xfade=transition=fade:duration=%.1f:offset=%.1f[outv]",
					transitionDuration, offsetTime),
				"-map", "[outv]",
				"-c:v", "libx264", "-preset", "medium", "-crf", "23",
				absNextOutput)

			output, err := xfadeCmd.CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("failed to create fade transition %d: %w, output: %s", i, err, string(output))
			}

			// Update current input for next iteration
			currentInput = absNextOutput
		}

		// Copy final temp file to combined video path
		if err := copyFile(currentInput, absCombinedVideoPath); err != nil {
			return "", fmt.Errorf("failed to copy final combined video: %w", err)
		}

		// Clean up temp files
		os.RemoveAll(tempDir)
	} else {
		// This shouldn't happen due to check at beginning, but just in case
		return "", fmt.Errorf("empty video segments array")
	}

	// At this point we have a combined video at combinedVideoPath regardless of which path we took above
	combinedVideoPath := filepath.Join(outputDir, "combined_video.mp4")
	absCombinedVideoPath, err := filepath.Abs(combinedVideoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for combined video: %w", err)
	}

	finalVideoPath := filepath.Join(outputDir, "final_video.mp4")
	absFinalVideoPath, err := filepath.Abs(finalVideoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for final video: %w", err)
	}

	// Add background music if provided
	if musicPath != "" && musicPath != "not-implemented" {
		// Get absolute path for music file
		absMusicPath, err := filepath.Abs(musicPath)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for music: %w", err)
		}

		// We already have absolute path for combinedVideoPath from earlier
		musicCmd := exec.Command("ffmpeg", "-i", absCombinedVideoPath, "-i", absMusicPath,
			"-map", "0:v", "-map", "1:a", "-shortest", "-c:v", "copy",
			"-c:a", "aac", "-b:a", "192k", absFinalVideoPath)

		output, err := musicCmd.CombinedOutput()
		if err != nil {
			// If adding music fails, just use the combined video
			log.Printf("Warning: Failed to add background music: %v, output: %s", err, string(output))
			// Copy the combined video as the final video
			if err := copyFile(absCombinedVideoPath, absFinalVideoPath); err != nil {
				return "", fmt.Errorf("failed to copy combined video: %w", err)
			}
		}
	} else {
		// If no music, just rename the combined video
		if err := copyFile(absCombinedVideoPath, absFinalVideoPath); err != nil {
			return "", fmt.Errorf("failed to copy combined video: %w", err)
		}
	}

	return absFinalVideoPath, nil
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

// getVideoDuration gets the duration of a video file in seconds using ffprobe
func getVideoDuration(videoPath string) float64 {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1", videoPath)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: Failed to get video duration: %v", err)
		return 5.0 // default duration if unable to determine
	}

	duration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		log.Printf("Warning: Failed to parse video duration: %v", err)
		return 5.0
	}

	return duration
}
