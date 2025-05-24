package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Copy of FoodAnalysisResult structure
type FoodAnalysisResult struct {
	Foods          []FoodItem       `json:"foods"`
	TotalNutrition NutritionSummary `json:"total_nutrition"`
	RawResponse    string           `json:"raw_response"`
}

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

type NutritionSummary struct {
	Calories string `json:"calories"`
	Fat      string `json:"fat"`
	Protein  string `json:"protein"`
	Carbs    string `json:"carbs"`
	Fiber    string `json:"fiber,omitempty"`
	Sodium   string `json:"sodium,omitempty"`
}

// Import the AnalyzeFoodImage function from parent directory
func AnalyzeFoodImage(imageURL string) (*FoodAnalysisResult, error) {
	// This is a mock implementation for testing
	// In real usage, you would call the actual function from the parent package

	// Create sample data for testing
	result := &FoodAnalysisResult{
		Foods: []FoodItem{
			{
				Name:     "Fried Chicken",
				Serving:  "8 pieces, ~300g",
				Calories: "500–600 kcal",
				Fat:      "25–35g (saturated 5–8g)",
				Protein:  "30–40g",
				Carbs:    "10–15g (from breading)",
				Sodium:   "1,200–1,500mg",
			},
			{
				Name:     "White Rice",
				Serving:  "1 cup, ~150g",
				Calories: "200–220 kcal",
				Fat:      "1–2g",
				Protein:  "4–5g",
				Carbs:    "45–50g",
				Fiber:    "1–2g",
			},
		},
		TotalNutrition: NutritionSummary{
			Calories: "1,600–1,850 kcal",
			Fat:      "60–75g",
			Protein:  "80–100g",
			Carbs:    "130–160g",
			Fiber:    "10–15g",
			Sodium:   "2,500–3,000mg",
		},
		RawResponse: "Sample raw response for testing",
	}

	return result, nil
}

func main() {
	// Test URL (this would be a real URL in production)
	testURL := "https://example.com/test-food-image.jpg"

	fmt.Println("Testing food analysis with URL:", testURL)
	result, err := AnalyzeFoodImage(testURL)

	if err != nil {
		log.Fatalf("Error analyzing food image: %v", err)
	}

	// Convert result to JSON for display
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("Error converting result to JSON: %v", err)
	}

	fmt.Println("Analysis Result:")
	fmt.Println(string(jsonData))

	// Save result to a file
	err = os.WriteFile("nutrition_test_result.json", jsonData, 0644)
	if err != nil {
		log.Printf("Error saving result to file: %v", err)
	} else {
		fmt.Println("Result saved to nutrition_test_result.json")
	}
}
