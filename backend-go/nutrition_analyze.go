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

	// The path to the Python script is one directory up from the backend-go directory
	scriptPath := filepath.Join(filepath.Dir(currentDir), "analysis.py")

	// Make sure the Python script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("analysis.py not found at %s: %w", scriptPath, err)
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

	// Set the working directory to where the script is
	cmd.Dir = filepath.Dir(scriptPath)

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
