# GoBiz Showcase

This application generates promotional videos for restaurants using AI. It includes a backend server built with Express.js and a frontend UI built with React and Chakra UI.

## Features

- Generate promotional videos for restaurants
- Customize restaurant details, opening and closing scenes
- Add multiple menu items with descriptions and images
- Configure background music settings
- Real-time progress tracking
- Video playback in the browser
- Automatic nutritional content analysis for food menu items

## Prerequisites

- Node.js v16 or higher
- npm or yarn
- FFmpeg (for video processing)

## Environment Variables

Create a `.env` file in the `backend` directory with the following variables:

```
OPENAI_API_KEY=your_openai_api_key
ALIBABA_API_KEY=your_alibaba_api_key
DASHSCOPE_API_KEY=your_dashscope_api_key
```

## Installation

### Backend

```bash
cd backend
npm install
```

### Frontend

```bash
cd frontend
npm install
```

## Running the Application

### Development Mode

To run both the backend server and frontend development server:

```bash
cd backend
npm run dev
```

This will start the backend server on port 3000 and the frontend development server on port 3001.

### Running Backend Only

```bash
cd backend
npm run server
```

### Running Frontend Only

```bash
cd frontend
npm run frontend
```

## Usage

1. Open your browser and navigate to http://localhost:3001
2. Fill out the restaurant details form
3. Add menu items with descriptions and images
4. Configure music settings if desired
5. Click "Generate Video" to start the process
6. Monitor the progress in real-time
7. Once complete, the video will be displayed for playback

## Nutrition Analysis Feature

The application now supports automatic nutritional content analysis for food menu items using AI. This feature helps restaurants provide nutritional information to their customers.

### How it works:

1. When a food image is added to a menu item, you can click the "Analyze Nutrition" button to get nutritional information
2. The system uses AI to analyze the food image and extract nutritional content
3. The nutritional information includes:
   - Calorie count
   - Protein content
   - Carbohydrates
   - Fat
   - Fiber (when available)
   - Sodium (when available)

### API Configuration:

To use the nutrition analysis feature, set the following environment variable:

```
DASHSCOPE_API_KEY=your_dashscope_api_key
```

This connects to Alibaba Cloud's DashScope API for image analysis.

## License

ISC
