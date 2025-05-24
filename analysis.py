#!/usr/bin/env python3
# filepath: /Users/ibrahim/Documents/alibaba2/analysis.py
from openai import OpenAI
import os
import sys
import argparse
import json


def get_response(image_url):
    client = OpenAI(
        api_key="sk-6326ac6e33024d8fa3abeeedef503abd",
        base_url="https://dashscope-intl.aliyuncs.com/compatible-mode/v1",
    )
    completion = client.chat.completions.create(
        model="qwen-vl-plus",
        messages=[
            {
              "role": "user",
              "content": [
                {
                  "type": "text",
                  "text": """Analyze this food image and output the result in ONLY valid JSON format exactly as follows, with no extra explanations or text outside the JSON. Make sure to identify all possible allergens that might be present in the food:

{
  "menu": "Menu item name",
  "description": "Brief description of the dish",
  "location": "Optional location/restaurant info if visible",
  "features": [
    "Feature 1 (e.g. MSG-Free)",
    "Feature 2 (e.g. calories info)"
  ],
  "foods_included": [
    {
      "name": "Category name (e.g. Vegetables)",
      "items": [
        "Food item 1",
        "Food item 2",
        "Food item 3"
      ]
    },
    {
      "name": "Category name (e.g. Protein)",
      "items": [
        "Food item 1",
        "Food item 2"
      ]
    }
  ],
  "ingredients": [
    "Ingredient 1",
    "Ingredient 2",
    "Ingredient 3"
  ],
  "allergens": [
    "Eggs",
    "Milk",
    "Peanuts",
    "Tree nuts",
    "Fish",
    "Shellfish",
    "Soy",
    "Wheat",
    "Gluten",
    "Sesame"
  ],
  "nutritional_content": {
    "calories": 400,
    "macronutrients": {
      "protein": {
        "amount": 25,
        "unit": "grams",
        "sources": ["source1", "source2"]
      },
      "carbohydrates": {
        "amount": 35,
        "unit": "grams",
        "sources": ["source1", "source2"]
      },
      "fat": {
        "amount": 15,
        "unit": "grams",
        "sources": ["source1", "source2"]
      }
    },
    "fiber": {
      "amount": 5,
      "unit": "grams",
      "sources": ["vegetables"]
    },
    "vitamins_minerals": {
      "vitamin_c": "Sources description",
      "vitamin_a": "Sources description"
    },
    "notes": [
      "Informational note 1",
      "Informational note 2"
    ]
  }
}"""
                },
                {
                  "type": "image_url",
                  "image_url": {
                    "url": image_url
                  }
                }
              ]
            }
          ]
        )
    
    # Parse the JSON response
    response_json = json.loads(completion.model_dump_json())
    
    # Extract and return just the content as plain text
    if 'choices' in response_json and len(response_json['choices']) > 0:
        content = response_json['choices'][0]['message']['content']
        
        # Try to ensure we're returning valid JSON
        try:
            # Try to parse the response as JSON to check validity
            food_json = json.loads(content)
            # Then re-format it with proper indentation
            return json.dumps(food_json, indent=2)
        except json.JSONDecodeError:
            # If the API didn't return proper JSON, try to extract any JSON-like content
            import re
            json_match = re.search(r'```json\s*(.*?)\s*```', content, re.DOTALL)
            if json_match:
                try:
                    food_json = json.loads(json_match.group(1))
                    return json.dumps(food_json, indent=2)
                except:
                    pass
            
            # If all parsing attempts fail, return the original content
            return content
    else:
        # If no choices found, return the whole response as a string
        # but with a clear error marker that processFoodAnalysis can handle
        return "ERROR PROCESSING IMAGE: " + json.dumps(response_json, indent=2)

if __name__=='__main__':
    parser = argparse.ArgumentParser(description='Analyze nutritional content in an image')
    parser.add_argument('image_url', help='URL of the image to analyze')
    
    args = parser.parse_args()
    
    try:
        result = get_response(args.image_url)
        # Print the raw result with no additional formatting
        # This will be captured by the Go code and passed to processFoodAnalysis
        print(result)
    except Exception as e:
        # Print error in a way the Go code can recognize and handle
        print(f"ERROR PROCESSING IMAGE: {e}", file=sys.stderr)
        sys.exit(1)