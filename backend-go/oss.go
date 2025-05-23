package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// uploadToOSS uploads a file to Alibaba Cloud OSS and returns the URL
func uploadToOSS(filePath string) (string, error) {
	// Get OSS configuration from environment variables
	ossRegion := os.Getenv("OSS_REGION")
	if ossRegion == "" {
		ossRegion = "oss-ap-southeast-5" // Default region
	}

	ossAccessKeyID := os.Getenv("OSS_ACCESS_KEY_ID")
	ossAccessKeySecret := os.Getenv("OSS_ACCESS_KEY_SECRET")
	ossBucket := os.Getenv("OSS_BUCKET")

	// Check if required credentials are present
	if ossAccessKeyID == "" || ossAccessKeySecret == "" || ossBucket == "" {
		return "", fmt.Errorf("OSS credentials not properly configured")
	}

	// Create OSS client
	endpoint := fmt.Sprintf("https://%s.aliyuncs.com", ossRegion)
	client, err := oss.New(endpoint, ossAccessKeyID, ossAccessKeySecret)
	if err != nil {
		return "", fmt.Errorf("failed to create OSS client: %w", err)
	}

	// Get bucket
	bucket, err := client.Bucket(ossBucket)
	if err != nil {
		return "", fmt.Errorf("failed to get bucket: %w", err)
	}

	// Define object key in OSS
	filename := filepath.Base(filePath)
	objectKey := fmt.Sprintf("images/%s", filename)

	// Upload file to OSS
	err = bucket.PutObjectFromFile(objectKey, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to OSS: %w", err)
	}

	// Generate URL for the uploaded file
	signedURL, err := bucket.SignURL(objectKey, oss.HTTPGet, 3600*24*365) // URL valid for 1 year
	if err != nil {
		// If we can't generate a signed URL, create a public URL
		return fmt.Sprintf("https://%s.%s/%s", ossBucket, endpoint, objectKey), nil
	}

	return signedURL, nil
}
