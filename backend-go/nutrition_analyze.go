package main

import (
	"encoding/json"
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
		"/home/ibrahim/alibaba2/analysis.py",                   // Root of project directory
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

	// Check if the output is in JSON format
	if strings.TrimSpace(rawOutput)[0] == '{' {
		// Try to parse as JSON first
		var foodData struct {
			Menu          string   `json:"menu"`
			Description   string   `json:"description"`
			Features      []string `json:"features"`
			Ingredients   []string `json:"ingredients"`
			Allergens     []string `json:"allergens"`
			FoodsIncluded []struct {
				Name  string   `json:"name"`
				Items []string `json:"items"`
			} `json:"foods_included"`
			NutritionalContent struct {
				Calories       int `json:"calories"`
				Macronutrients struct {
					Protein struct {
						Amount  interface{} `json:"amount"`
						Unit    string      `json:"unit"`
						Sources []string    `json:"sources"`
					} `json:"protein"`
					Carbohydrates struct {
						Amount  interface{} `json:"amount"`
						Unit    string      `json:"unit"`
						Sources []string    `json:"sources"`
					} `json:"carbohydrates"`
					Fat struct {
						Amount  interface{} `json:"amount"`
						Unit    string      `json:"unit"`
						Sources []string    `json:"sources"`
					} `json:"fat"`
				} `json:"macronutrients"`
			} `json:"nutritional_content"`
		}

		if err := json.Unmarshal([]byte(rawOutput), &foodData); err == nil {
			// Successfully parsed as JSON, convert to FoodAnalysisResult
			result := &FoodAnalysisResult{
				Foods:       make([]FoodItem, 0),
				RawResponse: rawOutput,
				Menu:        foodData.Menu,
				Description: foodData.Description,
				Ingredients: foodData.Ingredients,
				Allergens:   foodData.Allergens,
			}

			// Add the main item
			mainItem := FoodItem{
				Name:     foodData.Menu,
				Calories: fmt.Sprintf("%d kcal", foodData.NutritionalContent.Calories),
			}

			// Handle protein
			var proteinAmount string
			switch v := foodData.NutritionalContent.Macronutrients.Protein.Amount.(type) {
			case float64:
				proteinAmount = fmt.Sprintf("%.1f %s", v, foodData.NutritionalContent.Macronutrients.Protein.Unit)
			case int:
				proteinAmount = fmt.Sprintf("%d %s", v, foodData.NutritionalContent.Macronutrients.Protein.Unit)
			case string:
				proteinAmount = v
			default:
				proteinAmount = fmt.Sprintf("%v %s", v, foodData.NutritionalContent.Macronutrients.Protein.Unit)
			}
			mainItem.Protein = proteinAmount

			// Handle carbs
			var carbsAmount string
			switch v := foodData.NutritionalContent.Macronutrients.Carbohydrates.Amount.(type) {
			case float64:
				carbsAmount = fmt.Sprintf("%.1f %s", v, foodData.NutritionalContent.Macronutrients.Carbohydrates.Unit)
			case int:
				carbsAmount = fmt.Sprintf("%d %s", v, foodData.NutritionalContent.Macronutrients.Carbohydrates.Unit)
			case string:
				carbsAmount = v
			default:
				carbsAmount = fmt.Sprintf("%v %s", v, foodData.NutritionalContent.Macronutrients.Carbohydrates.Unit)
			}
			mainItem.Carbs = carbsAmount

			// Handle fat
			var fatAmount string
			switch v := foodData.NutritionalContent.Macronutrients.Fat.Amount.(type) {
			case float64:
				fatAmount = fmt.Sprintf("%.1f %s", v, foodData.NutritionalContent.Macronutrients.Fat.Unit)
			case int:
				fatAmount = fmt.Sprintf("%d %s", v, foodData.NutritionalContent.Macronutrients.Fat.Unit)
			case string:
				fatAmount = v
			default:
				fatAmount = fmt.Sprintf("%v %s", v, foodData.NutritionalContent.Macronutrients.Fat.Unit)
			}
			mainItem.Fat = fatAmount

			result.Foods = append(result.Foods, mainItem)

			// Set total nutrition to the same values since we have one item
			result.TotalNutrition = NutritionSummary{
				Calories: mainItem.Calories,
				Protein:  mainItem.Protein,
				Carbs:    mainItem.Carbs,
				Fat:      mainItem.Fat,
			}

			return result, nil
		}
	}

	// Fallback to the existing processFoodAnalysis function if JSON parsing fails
	return processFoodAnalysis(rawOutput), nil
}
