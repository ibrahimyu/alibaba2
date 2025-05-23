import express from 'express';
import cors from 'cors';
import { json } from 'body-parser';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { generateRestaurantVideo } from './script.js';
import progressTracker from './api/progress.js';
import crypto from 'crypto';

// Get directory name in ES modules
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Initialize express app
const app = express();
const PORT = process.env.PORT || 3000;

// Ensure output directory exists
const outputDir = path.join(__dirname, 'output');
if (!fs.existsSync(outputDir)) {
  fs.mkdirSync(outputDir, { recursive: true });
}

// Middleware
app.use(cors());
app.use(json({ limit: '50mb' }));
app.use('/output', express.static(outputDir));

// Route to handle video generation requests
app.post('/api/generate-video', async (req, res) => {
  try {
    const inputData = req.body;
    
    // Generate a unique job ID
    const jobId = crypto.randomUUID();
    
    // Save input data to a temporary JSON file
    const tempInputFile = path.join(__dirname, `input_${jobId}.json`);
    fs.writeFileSync(tempInputFile, JSON.stringify(inputData, null, 2));
    
    // Create a unique output directory for this request
    const outputDirName = `output_${jobId}`;
    const outputDir = path.join(__dirname, 'output', outputDirName);
    
    // Initialize progress tracking
    progressTracker.startJob(jobId);
    
    // Respond immediately with job ID for tracking
    res.json({
      success: true,
      message: 'Video generation started',
      jobId
    });
    
    try {
      // Start the video generation process asynchronously
      progressTracker.updateProgress(jobId, 'processing', 10, 'Processing input data');
      
      // Wrap the generateRestaurantVideo function to update progress
      const originalConsoleLog = console.log;
      console.log = (message, ...args) => {
        originalConsoleLog(message, ...args);
        
        // Update progress based on log messages
        if (typeof message === 'string') {
          if (message.includes('restaurant video generation')) {
            progressTracker.updateProgress(jobId, 'starting', 10, 'Starting video generation');
          } else if (message.includes('Generating video segments')) {
            progressTracker.updateProgress(jobId, 'segments', 20, 'Generating video segments');
          } else if (message.includes('Generating background music')) {
            progressTracker.updateProgress(jobId, 'music', 50, 'Generating background music');
          } else if (message.includes('Combining videos')) {
            progressTracker.updateProgress(jobId, 'combining', 80, 'Combining videos and adding music');
          } else if (message.includes('successfully generated at')) {
            progressTracker.updateProgress(jobId, 'finalizing', 95, 'Finalizing video');
          }
        }
      };
      
      const finalVideoPath = await generateRestaurantVideo(tempInputFile, outputDir);
      
      // Restore original console.log
      console.log = originalConsoleLog;
      
      // Clean up the temporary input file
      fs.unlinkSync(tempInputFile);
      
      // Construct the URL to access the video
      const videoUrl = `/output/${outputDirName}/${path.basename(finalVideoPath)}`;
      
      progressTracker.completeJob(jobId, `Video available at ${videoUrl}`);
    } catch (error) {
      console.error('Error generating video:', error);
      progressTracker.failJob(jobId, error instanceof Error ? error.message : String(error));
      
      // Don't need to send a response here as we've already responded above
    }
  } catch (error) {
    console.error('Error setting up video generation:', error);
    res.status(500).json({
      success: false,
      message: 'Failed to set up video generation',
      error: error instanceof Error ? error.message : String(error)
    });
  }
});

// Route to check generation progress for a specific job
app.get('/api/progress/:jobId', (req, res) => {
  const { jobId } = req.params;
  const progress = progressTracker.getJobProgress(jobId);
  res.json(progress);
});

// Route to get all active jobs
app.get('/api/jobs', (req, res) => {
  const jobs = progressTracker.getAllJobs();
  res.json(jobs);
});

// Start the server
app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});
