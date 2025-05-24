package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// VideoFormData represents the JSON structure for video generation
type VideoFormData struct {
	RestoName    string     `json:"resto_name"`
	RestoAddress string     `json:"resto_address"`
	OpeningScene Scene      `json:"opening_scene"`
	ClosingScene Scene      `json:"closing_scene"`
	Music        Music      `json:"music"`
	Menu         []MenuItem `json:"menu"`
}

// Scene represents a scene in the video
type Scene struct {
	Prompt   string `json:"prompt"`
	ImageURL string `json:"image_url"`
}

// Music represents the background music configuration
type Music struct {
	Enabled bool   `json:"enabled"`
	Genres  string `json:"genres"`
	BPM     int    `json:"bpm,omitempty"`
	Lyrics  string `json:"lyrics,omitempty"`
}

// MenuItem represents a menu item
type MenuItem struct {
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	PhotoURL    string `json:"photo_url"`
}

// JobProgress stores the progress of a video generation job
type JobProgress struct {
	JobID      string    `json:"job_id"`
	Status     string    `json:"status"` // processing, completed, failed
	Stage      string    `json:"stage"`
	Percent    int       `json:"percent"`
	Message    string    `json:"message"`
	VideoURL   string    `json:"video_url,omitempty"`
	Error      string    `json:"error,omitempty"`
	StartTime  time.Time `json:"start_time"`
	UpdateTime time.Time `json:"update_time"`
}

// Progress tracker
var (
	jobProgressMap = make(map[string]*JobProgress)
	progressMutex  sync.RWMutex
)

// setupAPIRoutes configures all API endpoints
func setupAPIRoutes(router fiber.Router, uploadsDir string) {
	// Image upload endpoint
	router.Post("/upload-image", func(c *fiber.Ctx) error {
		return handleImageUpload(c, uploadsDir)
	})

	// Video generation endpoint
	router.Post("/generate-video", func(c *fiber.Ctx) error {
		return handleVideoGeneration(c)
	})

	// Resume video generation endpoint
	router.Post("/resume-video/:jobId", func(c *fiber.Ctx) error {
		jobID := c.Params("jobId")

		// Check if job exists
		progressMutex.RLock()
		job, exists := jobProgressMap[jobID]
		progressMutex.RUnlock()

		if !exists {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Job not found",
			})
		}

		// Only failed jobs can be resumed
		if job.Status != "failed" {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "Only failed jobs can be resumed",
			})
		}

		// Reset job status
		progressMutex.Lock()
		job.Status = "processing"
		job.Error = ""
		job.Stage = "resuming"
		job.Message = "Resuming video generation"
		job.UpdateTime = time.Now()
		progressMutex.Unlock()

		// Create output directory path
		outputDirName := fmt.Sprintf("output_%s", jobID)
		outputDir := filepath.Join(".", "output", outputDirName)

		// Restore input data from checkpoint or request body
		var inputData VideoFormData
		checkpointInputFile := filepath.Join(os.TempDir(), fmt.Sprintf("input_%s.json", jobID))

		// If input file doesn't exist, use request body
		if _, err := os.Stat(checkpointInputFile); os.IsNotExist(err) {
			if err := c.BodyParser(&inputData); err != nil {
				return c.Status(400).JSON(fiber.Map{
					"success": false,
					"message": "Invalid request data",
					"error":   err.Error(),
				})
			}

			// Save the new input data
			inputJSON, _ := json.MarshalIndent(inputData, "", "  ")
			if err := os.WriteFile(checkpointInputFile, inputJSON, 0644); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"success": false,
					"message": "Failed to save input data",
					"error":   err.Error(),
				})
			}
		}

		// Start the video generation process asynchronously
		go func() {
			// Define progress callback for updating progress
			progressCallback := func(stage string, percent int, message string) {
				updateProgress(jobID, stage, percent, message)
			}

			// Call the actual video generation code (with checkpoint support)
			result, err := GenerateVideo(checkpointInputFile, outputDir, progressCallback)

			if err != nil {
				log.Printf("Error resuming video: %v", err)
				failJob(jobID, err.Error())
			} else {
				if result.Success {
					videoURL := fmt.Sprintf("/output/%s/%s", outputDirName, filepath.Base(result.VideoPath))
					completeJob(jobID, videoURL)
				} else {
					failJob(jobID, result.ErrorMessage)
				}
			}
		}()

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Video generation resumed",
			"jobId":   jobID,
		})
	})

	// Get job progress endpoint
	router.Get("/progress/:jobId", func(c *fiber.Ctx) error {
		jobID := c.Params("jobId")

		progressMutex.RLock()
		progress, exists := jobProgressMap[jobID]
		progressMutex.RUnlock()

		if !exists {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Job not found",
			})
		}

		return c.JSON(progress)
	})

	// Get all jobs endpoint
	router.Get("/jobs", func(c *fiber.Ctx) error {
		progressMutex.RLock()
		defer progressMutex.RUnlock()

		jobs := make([]*JobProgress, 0, len(jobProgressMap))
		for _, job := range jobProgressMap {
			jobs = append(jobs, job)
		}

		return c.JSON(jobs)
	})
}

// handleImageUpload processes uploaded images, resizes them, and uploads to OSS
func handleImageUpload(c *fiber.Ctx, uploadsDir string) error {
	// Get the file from the request
	file, err := c.FormFile("image")
	if err != nil {
		log.Println("Error getting file:", err)
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "No file uploaded",
		})
	}

	// Check if the file is an image
	contentType := file.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "File must be an image (JPEG, PNG, or GIF)",
		})
	}

	// Create a temporary file for the uploaded image
	tempFilename := fmt.Sprintf("%s-%s%s",
		time.Now().Format("20060102150405"),
		randomString(8),
		filepath.Ext(file.Filename))
	tempFilePath := filepath.Join(uploadsDir, tempFilename)

	// Save the uploaded file
	if err := c.SaveFile(file, tempFilePath); err != nil {
		log.Println("Error saving file:", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to save uploaded file",
			"error":   err.Error(),
		})
	}

	// Process the image (resize to 1280x720)
	resizedFilePath, err := resizeImage(tempFilePath)
	if err != nil {
		log.Println("Error resizing image:", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to process image",
			"error":   err.Error(),
		})
	}

	// Upload the image to OSS
	imageURL, err := uploadToOSS(resizedFilePath)
	if err != nil {
		log.Println("Error uploading to OSS:", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to upload image to OSS",
			"error":   err.Error(),
		})
	}

	// Clean up temporary files
	os.Remove(tempFilePath)
	if tempFilePath != resizedFilePath {
		os.Remove(resizedFilePath)
	}

	// Return the URL of the uploaded image
	return c.JSON(fiber.Map{
		"success": true,
		"url":     imageURL,
	})
}

// handleVideoGeneration processes video generation requests
func handleVideoGeneration(c *fiber.Ctx) error {
	// Parse input data
	var inputData VideoFormData
	if err := c.BodyParser(&inputData); err != nil {
		log.Println("Error parsing request body:", err)
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
	}

	// Generate a unique job ID
	jobID := uuid.New().String()

	// Create job progress tracker
	progressMutex.Lock()
	jobProgressMap[jobID] = &JobProgress{
		JobID:      jobID,
		Status:     "processing",
		Stage:      "processing",
		Percent:    10,
		Message:    "Processing input data",
		StartTime:  time.Now(),
		UpdateTime: time.Now(),
	}
	progressMutex.Unlock()

	// Save job to persistent storage
	go SaveJobs()

	// Create a unique output directory for this job
	outputDirName := fmt.Sprintf("output_%s", jobID)
	outputDir := filepath.Join(".", "output", outputDirName)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}

	// Save input data to a temporary JSON file
	tempInputFile := filepath.Join(os.TempDir(), fmt.Sprintf("input_%s.json", jobID))
	inputJSON, _ := json.MarshalIndent(inputData, "", "  ")
	if err := os.WriteFile(tempInputFile, inputJSON, 0644); err != nil {
		log.Println("Error writing input file:", err)
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to prepare input data",
			"error":   err.Error(),
		})
	}

	// Respond immediately with job ID for tracking
	c.JSON(fiber.Map{
		"success": true,
		"message": "Video generation started",
		"jobId":   jobID,
	})

	// Start the video generation process asynchronously
	go func() {
		// Define progress callback for updating progress
		progressCallback := func(stage string, percent int, message string) {
			updateProgress(jobID, stage, percent, message)
		}

		// Call the actual video generation code
		result, err := GenerateVideo(tempInputFile, outputDir, progressCallback)

		if err != nil {
			log.Printf("Error generating video: %v", err)
			failJob(jobID, err.Error())
		} else {
			if result.Success {
				// Get relative path for the UI
				videoURL := fmt.Sprintf("/output/%s/%s", outputDirName, filepath.Base(result.VideoPath))
				completeJob(jobID, videoURL)
			} else {
				failJob(jobID, result.ErrorMessage)
			}
		}

		// Clean up
		os.Remove(tempInputFile)
	}()

	return nil
}

// Helper functions for progress tracking
func updateProgress(jobID string, stage string, percent int, message string) {
	progressMutex.Lock()
	defer progressMutex.Unlock()

	if job, exists := jobProgressMap[jobID]; exists {
		job.Stage = stage
		job.Percent = percent
		job.Message = message
		job.UpdateTime = time.Now()

		// Save jobs to persistent storage
		go SaveJobs()
	}
}

func completeJob(jobID string, videoURL string) {
	progressMutex.Lock()
	defer progressMutex.Unlock()

	if job, exists := jobProgressMap[jobID]; exists {
		job.Status = "completed"
		job.Percent = 100
		job.Message = "Video generation complete"
		job.VideoURL = videoURL
		job.UpdateTime = time.Now()

		// Save jobs to persistent storage
		go SaveJobs()
	}
}

func failJob(jobID string, errorMsg string) {
	progressMutex.Lock()
	defer progressMutex.Unlock()

	if job, exists := jobProgressMap[jobID]; exists {
		job.Status = "failed"
		job.Error = errorMsg
		job.UpdateTime = time.Now()

		// Save jobs to persistent storage
		go SaveJobs()
	}
}

// Helper function to generate random string
func randomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// resizeImage resizes an image to 1280x720 resolution
func resizeImage(inputPath string) (string, error) {
	// Open the input image
	src, err := imaging.Open(inputPath)
	if err != nil {
		return "", err
	}

	// Resize the image while preserving aspect ratio and using a high-quality filter
	resized := imaging.Resize(src, 1280, 720, imaging.Lanczos)

	// Generate output filename
	outputFilename := fmt.Sprintf("resized_%s.jpg", randomString(8))
	outputPath := filepath.Join(filepath.Dir(inputPath), outputFilename)

	// Save the resized image
	err = imaging.Save(resized, outputPath, imaging.JPEGQuality(90))
	if err != nil {
		return "", err
	}

	return outputPath, nil
}
