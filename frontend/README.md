# Restaurant Promo Video Generator - Frontend

This is the React frontend for the Restaurant Promo Video Generator, built with Vite, React, TypeScript, and Chakra UI.

## Features

- Beautiful and intuitive UI for configuring restaurant promotional videos
- Image upload with automatic resizing to 1280x720 for video compatibility
- Support for custom image URLs as well as file uploads
- Real-time progress tracking during video generation
- Video preview after generation is complete
- Customizable music configuration

## Getting Started

1. Install dependencies:
   ```
   npm install
   ```

2. Start the development server:
   ```
   npm run dev
   ```

3. Build for production:
   ```
   npm run build
   ```

## Image Upload Feature

The application supports two ways to provide images for the video generation:

1. **Direct URL Entry**: Users can paste any public image URL into the input fields
2. **File Upload**: Users can upload images from their device

When a user uploads an image file:
1. The frontend displays a preview
2. The image is sent to the backend
3. The backend resizes it to 1280x720 resolution (optimal for video generation)
4. The image is stored on Alibaba OSS
5. The public URL from OSS is returned and automatically populated in the form field

This approach ensures all images used in video generation are properly formatted and accessible from the video generation API.

## API Integration

The frontend communicates with the Go backend through the following endpoints:

- `POST /api/upload-image` - Upload and resize images for video scenes and menu items
- `POST /api/generate-video` - Submit the video configuration for generation
- `GET /api/progress/:jobId` - Track the progress of video generation
- `GET /api/jobs` - List all video generation jobs

## Configuration Options

The video generator offers several customization options:

1. **Basic Info**: Restaurant name and address
2. **Opening Scene**: Description and image
3. **Menu Items**: Add multiple dishes with name, price, description, and image
4. **Closing Scene**: Description and image
5. **Music**: Enable background music generation with genre preferences and optional lyrics
