# Restaurant Promo Video Generator - Go Backend

This is the Go backend for the Restaurant Promo Video Generator, rewritten from Node.js using the Fiber framework.

## Features

- Image upload and processing (resize to 1280x720)
- Alibaba OSS integration for image storage
- Video generation using Alibaba Cloud's DashScope API
- Food video narration generation with OpenAI
- Background music generation with YuE AI
- Progress tracking for video generation
- RESTful API for frontend integration

## Prerequisites

- Go 1.24+
- Alibaba Cloud OSS account
- FFmpeg installed on the server
- Alibaba Cloud DashScope API key
- OpenAI API key (optional, for narration generation)
- YuE AI model repository (for music generation)

## Installation

1. Clone the repository
2. Install dependencies:
   ```
   go mod tidy
   ```
3. Set up environment variables in `.env` file:
   ```
   PORT=3000
   OSS_REGION=your-oss-region
   OSS_ACCESS_KEY_ID=your-access-key-id
   OSS_ACCESS_KEY_SECRET=your-access-key-secret
   OSS_BUCKET=your-bucket-name
   ALIBABA_API_KEY=your-dashscope-api-key
   OPENAI_API_KEY=your-openai-api-key
   YUE_REPO_PATH=/path/to/YuE
   YUE_CHECKPOINT_DIR=/path/to/YuE/checkpoints
   ```

## Running the Server

```
go run .
```

## API Routes

### Image Upload
- `POST /api/upload-image`
  - Request: multipart/form-data with "image" field
  - Response: `{ "success": true, "url": "image_url" }`

### Video Generation
- `POST /api/generate-video`
  - Request: JSON with video configuration
  - Response: `{ "success": true, "jobId": "unique_job_id" }`

### Progress Tracking
- `GET /api/progress/:jobId`
  - Response: Job progress details

### List All Jobs
- `GET /api/jobs`
  - Response: List of all job progress details

## YuE Music Generation Setup

The backend now integrates with YuE AI for background music generation. To set up YuE:

1. Clone the YuE repository:
   ```
   git clone https://github.com/multimodal-art-projection/YuE.git
   ```

2. Install the required dependencies for YuE (see full instructions in the YuE repository):
   ```
   # Create a conda environment
   conda create -n yue python=3.8
   conda activate yue
   
   # Install PyTorch with CUDA support
   conda install pytorch torchvision torchaudio cudatoolkit=11.8 -c pytorch -c nvidia
   
   # Install other dependencies
   pip install -r requirements.txt
   
   # Install FlashAttention 2 for memory optimization
   pip install flash-attn --no-build-isolation
   ```

3. Download the tokenizer:
   ```
   cd YuE/inference/
   git clone https://huggingface.co/m-a-p/xcodec_mini_infer
   ```

4. Set the environment variables in your `.env` file:
   ```
   YUE_REPO_PATH=/path/to/YuE
   YUE_CHECKPOINT_DIR=/path/to/YuE/checkpoints
   ```

5. The models will be automatically downloaded from Hugging Face when first used.

### Music Generation Options

You can customize the music generation by modifying the following parameters in the video generation request:

```json
{
  "music": {
    "enabled": true,
    "genres": "inspiring uplifting electronic bright vocal",
    "bpm": 120,
    "lyrics": "Your custom lyrics for the music generation"
  }
}
```

- `enabled`: Set to `true` to generate background music, `false` to skip
- `genres`: Space-separated tags describing the desired music style
- `bpm` (optional): Beats per minute for the music
- `lyrics` (optional): Custom lyrics for the music generation
