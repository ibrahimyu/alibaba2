package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// JobsDatabase represents the persistent storage of jobs
type JobsDatabase struct {
	Jobs      map[string]*JobProgress `json:"jobs"`
	UpdatedAt time.Time               `json:"updated_at"`
}

const jobsFilePath = "./data/jobs.json"

// SaveJobs saves all jobs to the jobs.json file
func SaveJobs() error {
	// Create database folder if it doesn't exist
	dbDir := filepath.Dir(jobsFilePath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return err
		}
	}

	// Lock the jobs map while we create a snapshot
	progressMutex.RLock()
	jobsSnapshot := make(map[string]*JobProgress)

	// Create a deep copy of each job to avoid race conditions
	for id, job := range jobProgressMap {
		jobCopy := *job // Create a copy of the job struct
		jobsSnapshot[id] = &jobCopy
	}
	progressMutex.RUnlock()

	// Create the database structure
	db := JobsDatabase{
		Jobs:      jobsSnapshot,
		UpdatedAt: time.Now(),
	}

	// Marshal to JSON
	jobsJSON, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return ioutil.WriteFile(jobsFilePath, jobsJSON, 0644)
}

// LoadJobs loads all jobs from the jobs.json file
func LoadJobs() error {
	// Check if file exists
	if _, err := os.Stat(jobsFilePath); os.IsNotExist(err) {
		log.Println("No jobs database found, starting fresh")
		return nil
	}

	// Read file
	data, err := ioutil.ReadFile(jobsFilePath)
	if err != nil {
		return err
	}

	// Parse JSON
	var db JobsDatabase
	if err := json.Unmarshal(data, &db); err != nil {
		return err
	}

	// Lock the jobs map and update it
	progressMutex.Lock()
	defer progressMutex.Unlock()

	// Clear existing jobs
	jobProgressMap = make(map[string]*JobProgress)

	// Add jobs from the database
	for id, job := range db.Jobs {
		jobProgressMap[id] = job
	}

	log.Printf("Loaded %d jobs from database", len(jobProgressMap))
	return nil
}
