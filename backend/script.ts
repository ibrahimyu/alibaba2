// Load the config first to ensure environment variables are available
import config from './api/config.js';

import fs from 'fs';
import path from 'path';
import { promisify } from 'util';
import { exec } from 'child_process';
import { generateFoodVideoNarration } from './api/narration.js';
import { backgroundMusic } from './api/bgm.js';

const execAsync = promisify(exec);

// Define interfaces for our input data
interface MenuItem {
  name: string;
  price: number;
  description: string;
  photo_url: string;
}

interface Scene {
  prompt: string;
  image_url: string;
}

interface Music {
  enabled: boolean;
  genres: string;
  bpm?: number;
  lyrics?: string;
}

interface RestaurantVideo {
  resto_name: string;
  resto_address: string;
  opening_scene: Scene;
  closing_scene: Scene;
  music: Music;
  menu: MenuItem[];
}

// Main function to process restaurant video
/**
 * Generate a restaurant promo video from JSON input
 * @param {string} inputFile - Path to JSON input file
 * @param {string} outputDir - Directory to save output video
 * @param {Function} [progressCallback] - Optional callback function for progress updates
 * @returns {Promise<string>} - Path to the generated video
 */
async function generateRestaurantVideo(
  inputFile: string, 
  outputDir: string,
  progressCallback?: (stage: string, percent: number, message: string) => void
) {
  try {
    console.log('Starting restaurant video generation...');
    
    // Call progress callback if provided
    if (progressCallback) {
      progressCallback('init', 5, 'Initializing video generation');
    }
    
    // Create output directory if it doesn't exist
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }
    
    const tempDir = path.join(outputDir, 'temp');
    if (!fs.existsSync(tempDir)) {
      fs.mkdirSync(tempDir, { recursive: true });
    }

    // Read and parse input JSON
    const inputData = JSON.parse(fs.readFileSync(inputFile, 'utf-8')) as RestaurantVideo;
    
    // Generate video segments in parallel
    console.log('Generating video segments...');
    const videoSegments = await generateVideoSegments(inputData, tempDir);
    
    // Generate background music if enabled
    let musicPath = '';
    if (inputData.music.enabled !== false) {
      console.log('Generating background music...');
      const musicResult = await generateBackgroundMusic(inputData, outputDir);
      musicPath = musicResult.outputPath as string;
    } else {
      console.log('Background music generation skipped (disabled in config)');
    }
    
    // Combine videos and add music if available
    console.log('Combining videos' + (musicPath ? ' and adding background music' : '') + '...');
    const finalVideo = await combineVideosWithMusic(
      videoSegments,
      musicPath,
      outputDir
    );
    
    console.log(`Restaurant video successfully generated at: ${finalVideo}`);
    return finalVideo;
    
  } catch (error) {
    console.error('Error generating restaurant video:', error);
    throw error;
  }
}

// Function to generate all video segments (opening, menu items, closing)
async function generateVideoSegments(
  inputData: RestaurantVideo, 
  tempDir: string
): Promise<string[]> {
  const videoSegmentPaths: string[] = [];
  const videoPromises: Promise<void>[] = [];
  
  // 1. Generate opening segment (5 seconds)
  videoPromises.push(
    (async () => {
      const openingVideoPath = await generatePromptVideo(
        inputData.opening_scene.prompt,
        inputData.opening_scene.image_url,
        'opening',
        tempDir
      );
      videoSegmentPaths.push(openingVideoPath);
    })()
  );
  
  // 2. Generate menu item segments (5 seconds each)
  for (let i = 0; i < inputData.menu.length; i++) {
    const menuItem = inputData.menu[i];
    videoPromises.push(
      (async () => {
        const menuNarration = await generateFoodVideoNarration(
          config.openAiApiKey,
          {
            foodName: menuItem.name,
            foodDescription: menuItem.description,
            narrationLength: 'short',
            toneOfVoice: 'enthusiastic'
          }
        );
        
        const menuVideoPath = await generatePromptVideo(
          menuNarration,
          menuItem.photo_url,
          `menu_${i}`,
          tempDir
        );
        
        // Store the path along with its index to maintain order
        videoSegmentPaths[i + 1] = menuVideoPath;
      })()
    );
  }
  
  // 3. Generate closing segment (5 seconds)
  videoPromises.push(
    (async () => {
      const closingVideoPath = await generatePromptVideo(
        inputData.closing_scene.prompt,
        inputData.closing_scene.image_url,
        'closing',
        tempDir
      );
      videoSegmentPaths.push(closingVideoPath);
    })()
  );
  
  // Wait for all video generation to complete
  await Promise.all(videoPromises);
  
  // Sort to ensure proper ordering (opening -> menu items -> closing)
  return videoSegmentPaths.filter(Boolean);
}

// Function to generate a single video segment with prompt and image
async function generatePromptVideo(
  prompt: string,
  imageUrl: string,
  segmentName: string,
  outputDir: string
): Promise<string> {
  try {
    console.log(`Generating video segment: ${segmentName}`);
    
    // Generate video using Alibaba API from our config singleton
    const videoResponse = await config.alibabaApi.generateVideoFromImage({
      prompt,
      imageInput: imageUrl,
      imageType: 'url',
      resolution: '720P',
      promptExtend: true
    });
    
    // Get task ID and check progress
    const taskId = videoResponse.output.task_id;
    console.log(`Video task started for segment ${segmentName}, task ID: ${taskId}`);
    
    // Poll for completion
    const result = await config.alibabaApi.pollTaskCompletion(taskId);
    
    if (result.output.task_status !== 'SUCCEEDED') {
      throw new Error(`Video generation failed for segment ${segmentName}: ${result.output.message}`);
    }
    
    // Download the video file
    const videoUrl = result.output.video_url as string;
    const videoPath = path.join(outputDir, `${segmentName}.mp4`);
    
    await downloadFile(videoUrl, videoPath);
    
    return videoPath;
  } catch (error) {
    console.error(`Error generating video segment ${segmentName}:`, error);
    throw error;
  }
}

// Function to generate background music based on restaurant theme
async function generateBackgroundMusic(
  inputData: RestaurantVideo, 
  outputDir: string
): Promise<{ outputPath: string }> {
  // Extract genre tags
  let genreTags = inputData.music.genres || 'ambient instrumental lounge';
  
  // Add BPM tag if specified
  if (inputData.music.bpm) {
    genreTags += ` ${inputData.music.bpm}bpm`;
  }
  
  // Create lyrics content
  const lyrics = inputData.music.lyrics ? 
    `[instrumental]\n${inputData.music.lyrics}` :
    `[instrumental]\n${inputData.resto_name} - ${inputData.resto_address}\n${inputData.opening_scene.prompt}`;
  
  // Generate music using YuE
  console.log(`Generating background music with genre: ${genreTags}`);
  const musicResult = await backgroundMusic.generate({
    genre: genreTags,
    lyrics: lyrics,
    outputDir: path.join(outputDir, 'music'),
    runNSegments: 2
  });
  
  if (!musicResult.success || !musicResult.outputPath) {
    throw new Error('Failed to generate background music');
  }
  
  return { outputPath: musicResult.outputPath };
}

// Function to combine all video segments and add background music
async function combineVideosWithMusic(
  videoSegmentPaths: string[],
  musicPath: string,
  outputDir: string
): Promise<string> {
  // Create file list for FFmpeg
  const fileListPath = path.join(outputDir, 'filelist.txt');
  const fileListContent = videoSegmentPaths
    .map(videoPath => `file '${videoPath}'`)
    .join('\n');
  
  fs.writeFileSync(fileListPath, fileListContent);
  
  // Combine videos
  const combinedVideoPath = path.join(outputDir, 'combined_video.mp4');
  await execAsync(`ffmpeg -f concat -safe 0 -i ${fileListPath} -c copy ${combinedVideoPath}`);
  
  // Add background music if provided
  const finalVideoPath = path.join(outputDir, 'final_video.mp4');
  
  if (musicPath) {
    await execAsync(`ffmpeg -i ${combinedVideoPath} -i ${musicPath} -map 0:v -map 1:a -shortest -c:v copy -c:a aac -b:a 192k ${finalVideoPath}`);
  } else {
    // If no music, just rename the combined video
    fs.copyFileSync(combinedVideoPath, finalVideoPath);
  }
  
  return finalVideoPath;
}

// Helper function to download a file from a URL
async function downloadFile(url: string, destPath: string): Promise<void> {
  const response = await fetch(url);
  const arrayBuffer = await response.arrayBuffer();
  fs.writeFileSync(destPath, Buffer.from(arrayBuffer));
}

// Helper function to adjust video duration using FFmpeg
async function adjustVideoDuration(
  inputPath: string,
  outputPath: string,
  durationSeconds: number
): Promise<void> {
  await execAsync(`ffmpeg -i ${inputPath} -t ${durationSeconds} -c:v copy -c:a copy ${outputPath}`);
}

// Function to run the video generation with settings from command line or environment variables
async function main() {
  // Get input file and output directory from command line arguments or use defaults
  const inputFile = process.argv[2] || './sample_input_new.json';
  const outputDir = process.argv[3] || './output';
  
  try {
    const finalVideoPath = await generateRestaurantVideo(
      inputFile,
      outputDir
    );
    console.log(`Restaurant promo video generated successfully: ${finalVideoPath}`);
  } catch (error) {
    console.error('Failed to generate restaurant video:', error);
    process.exit(1);
  }
}

// Run the main function if this script is executed directly
if (require.main === module) {
  main();
}

export {
  generateRestaurantVideo
};