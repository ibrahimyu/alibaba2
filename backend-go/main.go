package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, using environment variables")
	}

	// Create a new Fiber instance
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB limit for file uploads
	})

	// Add middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Ensure output directory exists
	outputDir := filepath.Join(".", "output")
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}

	// Ensure uploads directory exists
	uploadsDir := filepath.Join(".", "uploads")
	if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
		os.MkdirAll(uploadsDir, 0755)
	}

	// Serve static files from output directory
	app.Static("/output", outputDir)

	// Setup routes
	setupRoutes(app, uploadsDir)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Server listening on port %s\n", port)
	log.Fatal(app.Listen(":" + port))
}

func setupRoutes(app *fiber.App, uploadsDir string) {
	// Health check route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// Register other API routes
	api := app.Group("/api")
	setupAPIRoutes(api, uploadsDir)
}
