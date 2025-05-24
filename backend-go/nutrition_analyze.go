package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func AnalyzeFoodImage2(ImageURL string) (*FoodAnalysisResult, error) {
	// Get the current directory
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check multiple possible locations for the Python script
	possibleLocations := []string{
		filepath.Join(filepath.Dir(currentDir), "analysis.py"), // Parent directory
		filepath.Join(currentDir, "analysis.py"),               // Current directory
		"/Users/ibrahim/Documents/alibaba2/analysis.py",        // Root of project directory
	}

	// Find the first location that exists
	scriptPath := ""
	for _, path := range possibleLocations {
		if _, err := os.Stat(path); err == nil {
			scriptPath = path
			fmt.Printf("Found analysis.py at: %s\n", scriptPath)
			break
		}
	}

	// Make sure the Python script exists
	if scriptPath == "" {
		return nil, fmt.Errorf("analysis.py not found in any of these locations: %v", possibleLocations)
	}

	// Check if there's a virtual environment in the project directory
	venvPath := filepath.Join(filepath.Dir(scriptPath), "venv")
	venvExists := false
	if _, err := os.Stat(venvPath); err == nil {
		venvExists = true
	}

	var cmd *exec.Cmd
	if venvExists {
		// If on Unix-like system (macOS/Linux), use the virtual environment's Python
		venvPython := filepath.Join(venvPath, "bin", "python")
		if _, err := os.Stat(venvPython); err == nil {
			cmd = exec.Command(venvPython, scriptPath, ImageURL)
		} else {
			// Fallback to system Python
			cmd = exec.Command("python3", scriptPath, ImageURL)
		}
	} else {
		// Use system Python if no venv
		cmd = exec.Command("python3", scriptPath, ImageURL)
	}

	// Set the working directory to the parent directory of the script
	// This ensures any relative paths in the script work correctly
	scriptDir := filepath.Dir(scriptPath)
	cmd.Dir = scriptDir
	fmt.Printf("Setting working directory to: %s\n", scriptDir)

	// Add debug information to environment
	cmd.Env = os.Environ()

	// Print the command we're about to run for debugging
	fmt.Printf("Executing command: %s %s\n", cmd.Path, strings.Join(cmd.Args[1:], " "))

	// Capture the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running Python script: %w, output: %s", err, string(output))
	}

	// The output of the script is the text content of the nutrition analysis
	rawOutput := string(output)

	fmt.Println("Python script result:")
	fmt.Println(rawOutput)

	// Check if the output indicates an error
	if strings.Contains(rawOutput, "ERROR PROCESSING IMAGE:") {
		return nil, fmt.Errorf("python script error: %s", rawOutput)
	}

	// Process the result using the existing processFoodAnalysis function
	return processFoodAnalysis(rawOutput), nil
}
