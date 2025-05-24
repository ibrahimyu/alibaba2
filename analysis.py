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
                  "text": "What is the nutritional content with accurate numbers in this picture and what foods are in it and output ingredients and nutrition only, no explanations"
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
        return response_json['choices'][0]['message']['content']
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