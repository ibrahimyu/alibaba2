// filepath: /Users/ibrahim/Documents/alisandbox/api/bgm.ts

import fs from 'fs';
import path from 'path';
import { promisify } from 'util';
import { exec } from 'child_process';

const execAsync = promisify(exec);

/**
 * Interface for YuE background music generation parameters
 */
interface YueBgmGenerationParams {
  // Required parameters
  genre: string;         // Music genre (e.g., "inspiring electronic bright vocal")
  lyrics?: string;       // Optional lyrics for vocal music (if not provided, instrumental only)
  outputDir: string;     // Directory to store the generated music
  
  // Model and generation settings
  stage1Model?: string;  // Stage 1 model name (default: "m-a-p/YuE-s1-7B-anneal-en-cot")
  stage2Model?: string;  // Stage 2 model name (default: "m-a-p/YuE-s2-1B-general")
  runNSegments?: number; // Number of segments to generate (default: 2)
  stage2BatchSize?: number; // Batch size for stage 2 (default: 4)
  maxNewTokens?: number; // Maximum number of tokens to generate (default: 3000)
  repetitionPenalty?: number; // Repetition penalty (default: 1.1)
  
  // Reference audio for style (optional)
  referenceAudio?: {
    audioPath?: string;         // For single-track ICL
    vocalTrackPath?: string;    // For dual-track ICL
    instrumentalTrackPath?: string; // For dual-track ICL
    startTime?: number;         // Start time in seconds (default: 0)
    endTime?: number;           // End time in seconds (default: 30)
  };

  // GPU settings
  cudaIdx?: number;      // CUDA device index (default: 0)
}

/**
 * Interface for the results of the YuE music generation process
 */
interface YueBgmGenerationResult {
  success: boolean;
  outputPath?: string;      // Path to the generated music file
  mixPath?: string;         // Path to the mixed audio file
  vocalPath?: string;       // Path to the vocal track
  instrumentalPath?: string; // Path to the instrumental track
  error?: string;           // Error message if generation failed
  logs?: string;            // Generation process logs
}

/**
 * Generate background music for videos using YuE model
 * 
 * @param params - Parameters for the background music generation
 * @returns Promise with the results of the music generation
 */
export async function generateBackgroundMusic(
  params: YueBgmGenerationParams
): Promise<YueBgmGenerationResult> {
  try {
    // Create a temporary directory for YuE files
    const tmpDir = path.join(params.outputDir, 'yue_tmp');
    if (!fs.existsSync(tmpDir)) {
      fs.mkdirSync(tmpDir, { recursive: true });
    }

    // Create a unique identifier for this generation
    const timestamp = Date.now();
    const runId = `bgm_${timestamp}`;
    
    // Create genre.txt file
    const genreTxtPath = path.join(tmpDir, `genre_${runId}.txt`);
    fs.writeFileSync(genreTxtPath, params.genre);
    
    // Create lyrics.txt file (if lyrics provided)
    const lyricsTxtPath = path.join(tmpDir, `lyrics_${runId}.txt`);
    
    if (params.lyrics) {
      // Format lyrics properly for YuE
      const formattedLyrics = formatYueLyrics(params.lyrics);
      fs.writeFileSync(lyricsTxtPath, formattedLyrics);
    } else {
      // For instrumental only, we create a simple lyrics file with [instrumental] tag
      fs.writeFileSync(lyricsTxtPath, "[instrumental]");
    }

    // Build the YuE command
    const command = buildYueCommand(runId, tmpDir, params);
    
    // Execute the YuE command
    console.log('Starting YuE background music generation...');
    const { stdout, stderr } = await execAsync(command);
    
    // Find the generated files
    const outputFiles = findOutputFiles(params.outputDir, runId);
    
    if (!outputFiles.mixPath) {
      throw new Error('Music generation failed: No output files found');
    }
    
    return {
      success: true,
      ...outputFiles,
      logs: stdout
    };
  } catch (error) {
    console.error('Error generating background music:', error);
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Unknown error occurred'
    };
  }
}

/**
 * Format lyrics for YuE model input
 * @param rawLyrics - The raw lyrics text
 * @returns Properly formatted lyrics for YuE
 */
function formatYueLyrics(rawLyrics: string): string {
  // Split the lyrics into lines
  const lines = rawLyrics.split('\n').filter(line => line.trim().length > 0);
  
  // Check if there are already section markers like [verse], [chorus]
  const hasStructureLabels = lines.some(line => 
    /^\[(verse|chorus|bridge|outro|intro)\]/i.test(line.trim())
  );
  
  if (hasStructureLabels) {
    // Already formatted correctly
    return rawLyrics;
  }
  
  // Simple lyrics structuring: create verse and chorus
  const halfwayPoint = Math.floor(lines.length / 2);
  
  const verseLines = lines.slice(0, halfwayPoint);
  const chorusLines = lines.slice(halfwayPoint);
  
  return `[verse]\n${verseLines.join('\n')}\n\n[chorus]\n${chorusLines.join('\n')}`;
}

/**
 * Build the YuE command line based on the provided parameters
 */
function buildYueCommand(runId: string, tmpDir: string, params: YueBgmGenerationParams): string {
  const {
    outputDir,
    stage1Model = "m-a-p/YuE-s1-7B-anneal-en-cot",
    stage2Model = "m-a-p/YuE-s2-1B-general",
    runNSegments = 2,
    stage2BatchSize = 4,
    maxNewTokens = 3000,
    repetitionPenalty = 1.1,
    cudaIdx = 0,
    referenceAudio
  } = params;

  const genreTxtPath = path.join(tmpDir, `genre_${runId}.txt`);
  const lyricsTxtPath = path.join(tmpDir, `lyrics_${runId}.txt`);
  
  // Base command
  let cmd = `cd $HOME && cd YuE/inference/ && python infer.py \\
    --cuda_idx ${cudaIdx} \\
    --stage1_model ${stage1Model} \\
    --stage2_model ${stage2Model} \\
    --genre_txt ${genreTxtPath} \\
    --lyrics_txt ${lyricsTxtPath} \\
    --run_n_segments ${runNSegments} \\
    --stage2_batch_size ${stage2BatchSize} \\
    --output_dir ${outputDir} \\
    --max_new_tokens ${maxNewTokens} \\
    --repetition_penalty ${repetitionPenalty}`;
  
  // Add reference audio parameters if provided
  if (referenceAudio) {
    const startTime = referenceAudio.startTime || 0;
    const endTime = referenceAudio.endTime || 30;
    
    // Check if we're using dual-track or single-track ICL
    if (referenceAudio.vocalTrackPath && referenceAudio.instrumentalTrackPath) {
      // Use dual-track ICL mode
      cmd += ` \\
    --use_dual_tracks_prompt \\
    --vocal_track_prompt_path ${referenceAudio.vocalTrackPath} \\
    --instrumental_track_prompt_path ${referenceAudio.instrumentalTrackPath} \\
    --prompt_start_time ${startTime} \\
    --prompt_end_time ${endTime}`;
      
      // Switch to ICL model if using default CoT model
      if (stage1Model === "m-a-p/YuE-s1-7B-anneal-en-cot") {
        cmd = cmd.replace("m-a-p/YuE-s1-7B-anneal-en-cot", "m-a-p/YuE-s1-7B-anneal-en-icl");
      }
    } else if (referenceAudio.audioPath) {
      // Use single-track ICL mode
      cmd += ` \\
    --use_audio_prompt \\
    --audio_prompt_path ${referenceAudio.audioPath} \\
    --prompt_start_time ${startTime} \\
    --prompt_end_time ${endTime}`;
      
      // Switch to ICL model if using default CoT model
      if (stage1Model === "m-a-p/YuE-s1-7B-anneal-en-cot") {
        cmd = cmd.replace("m-a-p/YuE-s1-7B-anneal-en-cot", "m-a-p/YuE-s1-7B-anneal-en-icl");
      }
    }
  }
  
  return cmd;
}

/**
 * Find the generated output files from YuE
 */
function findOutputFiles(outputDir: string, runId: string): {
  mixPath?: string;
  vocalPath?: string;
  instrumentalPath?: string;
  outputPath?: string;
} {
  // Look for files in the output directory
  try {
    const files = fs.readdirSync(outputDir);
    
    // YuE output format is not specified in the documentation, so we're making assumptions:
    // We'll look for the most recently created .mp3 files
    const mp3Files = files.filter(f => f.endsWith('.mp3'));
    
    if (mp3Files.length === 0) {
      return {};
    }
    
    // Sort by creation time, most recent first
    mp3Files.sort((a, b) => {
      const statA = fs.statSync(path.join(outputDir, a));
      const statB = fs.statSync(path.join(outputDir, b));
      return statB.mtimeMs - statA.mtimeMs;
    });
    
    // Try to identify the different tracks
    const mixFile = mp3Files.find(f => !f.includes('Vocals') && !f.includes('Instrumental'));
    const vocalFile = mp3Files.find(f => f.includes('Vocals'));
    const instrumentalFile = mp3Files.find(f => f.includes('Instrumental'));
    
    return {
      outputPath: mixFile ? path.join(outputDir, mixFile) : undefined,
      mixPath: mixFile ? path.join(outputDir, mixFile) : undefined,
      vocalPath: vocalFile ? path.join(outputDir, vocalFile) : undefined,
      instrumentalPath: instrumentalFile ? path.join(outputDir, instrumentalFile) : undefined
    };
  } catch (error) {
    console.error('Error finding output files:', error);
    return {};
  }
}

/**
 * Generate instrumental-only background music (no vocals)
 * 
 * @param params - Parameters for BGM generation except lyrics
 * @returns Promise with the results of the music generation
 */
export async function generateInstrumentalBgm(
  params: Omit<YueBgmGenerationParams, 'lyrics'>
): Promise<YueBgmGenerationResult> {
  return generateBackgroundMusic({
    ...params,
    lyrics: undefined // No lyrics for instrumental
  });
}

/**
 * Higher-level wrapper for different types of background music generation
 */
export const backgroundMusic = {
  generate: generateBackgroundMusic,
  generateInstrumental: generateInstrumentalBgm,
};

export default backgroundMusic;