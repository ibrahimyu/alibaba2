// Import the config singleton for access to API keys
import config from './config.js';

/**
 * This module provides utility functions for generating video content,
 * including text narration generation using OpenAI-compatible APIs.
 */

/**
 * Response structure for OpenAI API completion
 */
interface OpenAIResponse {
  id: string;
  object: string;
  created: number;
  model: string;
  choices: {
    index: number;
    message: {
      role: string;
      content: string;
    };
    finish_reason: string;
  }[];
  usage: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}

/**
 * Parameters for generating food video narration
 */
interface FoodNarrationParams {
  foodName: string;
  foodDescription: string;
  narrationLength?: 'short' | 'medium' | 'long';
  toneOfVoice?: string;
  targetAudience?: string;
  creativityLevel?: number; // 0 to 1, controls temperature
}

/**
 * Parameters for generating food song lyrics
 */
interface FoodLyricsParams {
  foodName: string;
  foodDescription: string;
  songStyle?: string;         // e.g., "pop", "rock", "jazz", "folk", "rap"
  verseCount?: number;        // Number of verses to generate (default: 2)
  includeChorus?: boolean;    // Whether to include a chorus (default: true)
  mood?: string;              // e.g., "upbeat", "melancholic", "inspirational"
  creativityLevel?: number;   // 0 to 1, controls temperature
}

/**
 * Default system prompt for the video narration request
 */
const DEFAULT_SYSTEM_PROMPT = 
`You are a professional food video script writer. 
Your task is to create compelling, engaging, and concise video narrations about food.
Focus on sensory language, vivid descriptions, and evocative content that makes viewers crave the food.
Keep the narration conversational, engaging, and suitable for video content.`;

/**
 * Default system prompt for the song lyrics request
 */
const LYRICS_SYSTEM_PROMPT = 
`You are a professional songwriter specializing in creating catchy, memorable lyrics about food.
Your task is to create engaging, creative song lyrics that highlight the qualities and emotions associated with the food.
The lyrics should be well-structured with clear sections (verse, chorus, etc.) and maintain a consistent rhythm and rhyme scheme.
Your lyrics should be evocative and make people excited about the food being described.`;

/**
 * Generate a video narration script for food using an OpenAI-compatible API
 * 
 * @param apiKey - OpenAI or compatible API key (Optional, uses config.openAiApiKey if not provided)
 * @param params - Parameters for narration generation 
 * @param endpoint - API endpoint (defaults to OpenAI)
 * @returns The generated narration text
 */
export async function generateFoodVideoNarration(
  apiKey: string = config.alibabaApiKey,
  params: FoodNarrationParams,
  endpoint: string = 'https://dashscope-intl.aliyuncs.com/compatible-mode/v1'
): Promise<string> {
  const { 
    foodName,
    foodDescription,
    narrationLength = 'short',
    toneOfVoice = 'enthusiastic',
    targetAudience = 'general',
    creativityLevel = 0.7
  } = params;

  // Define length constraints based on narrationLength
  const wordCount = narrationLength === 'short' 
    ? '30-50 words' 
    : narrationLength === 'medium' 
      ? '80-120 words' 
      : '150-200 words';

  // Construct the user prompt
  const userPrompt = `Generate a ${narrationLength} video narration script (${wordCount}) for "${foodName}".
Food description: ${foodDescription}
Tone: ${toneOfVoice}
Target audience: ${targetAudience}
The narration should be engaging, use sensory language, and make viewers crave the food.
Do NOT include any camera directions or technical instructions.
Return ONLY the narration text that would be spoken in the video.`;

  try {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${apiKey}`
      },
      body: JSON.stringify({
        model: 'gpt-4o', // Using gpt-4o, adjust as needed for compatibility
        messages: [
          {
            role: 'system',
            content: DEFAULT_SYSTEM_PROMPT
          },
          {
            role: 'user',
            content: userPrompt
          }
        ],
        temperature: creativityLevel,
        max_tokens: 500
      })
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`API request failed with status ${response.status}: ${errorText}`);
    }

    const data = await response.json() as OpenAIResponse;
    return data.choices[0].message.content.trim();
  } catch (error) {
    console.error('Error generating food video narration:', error);
    throw error;
  }
}

/**
 * Generate song lyrics about food using an OpenAI-compatible API
 * 
 * @param apiKey - OpenAI or compatible API key (Optional, uses config.openAiApiKey if not provided)
 * @param params - Parameters for lyrics generation
 * @param endpoint - API endpoint (defaults to OpenAI)
 * @returns The generated song lyrics with structure markers
 */
export async function generateFoodSongLyrics(
  apiKey: string = config.openAiApiKey,
  params: FoodLyricsParams,
  endpoint: string = 'https://api.openai.com/v1/chat/completions'
): Promise<string> {
  const {
    foodName,
    foodDescription,
    songStyle = 'pop',
    verseCount = 2,
    includeChorus = true,
    mood = 'upbeat',
    creativityLevel = 0.8
  } = params;

  // Construct the user prompt
  const userPrompt = `Write ${verseCount} verse(s) ${includeChorus ? 'and a chorus ' : ''}for a ${mood} ${songStyle} song about "${foodName}".

Food description: ${foodDescription}

Requirements:
- Structure the lyrics with clear section markers like [verse], [chorus], etc.
- The lyrics should have a consistent rhythm and rhyme scheme appropriate for ${songStyle} music
- Make the lyrics catchy, memorable, and enjoyable to sing along with
- Focus on sensory language that makes the food appealing
- Keep each verse to 4-8 lines, and the chorus (if included) to 4-6 lines
- Ensure the lyrics would work well with the YuE music generation system
- Each section should be separated by two newline characters

Return ONLY the structured lyrics without any explanations.`;

  try {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${apiKey}`
      },
      body: JSON.stringify({
        model: 'gpt-4o', // Using gpt-4o, adjust as needed for compatibility
        messages: [
          {
            role: 'system',
            content: LYRICS_SYSTEM_PROMPT
          },
          {
            role: 'user',
            content: userPrompt
          }
        ],
        temperature: creativityLevel,
        max_tokens: 800
      })
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`API request failed with status ${response.status}: ${errorText}`);
    }

    const data = await response.json() as OpenAIResponse;
    const lyrics = data.choices[0].message.content.trim();
    
    // Ensure proper formatting for YuE
    return formatLyricsForYuE(lyrics);
  } catch (error) {
    console.error('Error generating food song lyrics:', error);
    throw error;
  }
}

/**
 * Format the generated lyrics to ensure compatibility with YuE
 * 
 * @param lyrics - The raw generated lyrics
 * @returns Properly formatted lyrics for YuE
 */
function formatLyricsForYuE(lyrics: string): string {
  // Ensure section markers are properly formatted
  let formattedLyrics = lyrics
    // Normalize section markers to proper format
    .replace(/\[([^\]]+)\]/g, (match, p1) => {
      const section = p1.toLowerCase().trim();
      if (section === 'verse' || section.startsWith('verse ')) {
        return '[verse]';
      } else if (section === 'chorus' || section.startsWith('chorus ')) {
        return '[chorus]';
      } else if (section === 'bridge') {
        return '[bridge]';
      } else if (section === 'outro') {
        return '[outro]';
      } else if (section === 'intro') {
        return '[intro]';
      }
      // Default to verse if unknown section
      return '[verse]';
    })
    // Ensure double newlines between sections
    .replace(/\[verse\]/g, '\n\n[verse]')
    .replace(/\[chorus\]/g, '\n\n[chorus]')
    .replace(/\[bridge\]/g, '\n\n[bridge]')
    .replace(/\[outro\]/g, '\n\n[outro]')
    .replace(/\[intro\]/g, '\n\n[intro]');

  // Remove any leading newlines
  formattedLyrics = formattedLyrics.replace(/^\n+/, '');
  
  // Ensure first section has no leading newlines
  if (!formattedLyrics.startsWith('[')) {
    formattedLyrics = '[verse]\n' + formattedLyrics;
  }

  return formattedLyrics;
}

/**
 * Generate a video narration and combine it with food images to create a video
 * using both LLM for narration and Alibaba API for video generation
 * 
 * @param apiConfig - Configuration for API access (optional, uses config singleton if not provided)
 * @param foodData - Food information for video generation
 * @returns Object containing narration text and task ID for the video generation
 */
export async function createFoodVideoWithNarration(
  apiConfig: {
    openAiKey?: string;
    alibabaKey?: string;
    openAiEndpoint?: string;
  } = {},
  foodData: {
    name: string;
    description: string;
    imageUrls: string[];
    narrationOptions?: Partial<FoodNarrationParams>;
  }
): Promise<{ narration: string; taskId: string }> {
  // First, generate the narration
  const narration = await generateFoodVideoNarration(
    apiConfig.openAiKey || config.openAiApiKey,
    {
      foodName: foodData.name,
      foodDescription: foodData.description,
      ...(foodData.narrationOptions || {})
    },
    apiConfig.openAiEndpoint
  );

  // Generate video using the narration and provided images
  const videoResponse = await config.alibabaApi.generateVideoFromMultipleImages({
    prompt: narration,  // Use the generated narration as the prompt
    refImagesUrls: foodData.imageUrls,
    size: '1280*720'
  });

  return {
    narration,
    taskId: videoResponse.output.task_id
  };
}

/**
 * Generate song lyrics and music for a food video using YuE
 * 
 * @param apiConfig - Configuration for API access (optional, uses config singleton if not provided)
 * @param foodData - Food information for music generation
 * @returns Object containing lyrics, music file paths, and task ID
 */
export async function createFoodSongWithLyrics(
  apiConfig: {
    openAiKey?: string;
    yueConfig?: {
      outputDir: string;
      stage1Model?: string;
      stage2Model?: string;
    }
  } = {},
  foodData: {
    name: string;
    description: string;
    songStyle?: string;
    mood?: string;
    lyricsOptions?: Partial<FoodLyricsParams>;
    musicOptions?: {
      referenceAudio?: {
        audioPath?: string;
        vocalTrackPath?: string;
        instrumentalTrackPath?: string;
        startTime?: number;
        endTime?: number;
      };
    }
  }
): Promise<{ lyrics: string; musicResult: any }> {
  // First, generate the lyrics
  const lyrics = await generateFoodSongLyrics(
    apiConfig.openAiKey || config.openAiApiKey,
    {
      foodName: foodData.name,
      foodDescription: foodData.description,
      songStyle: foodData.songStyle || 'pop',
      mood: foodData.mood || 'upbeat',
      ...(foodData.lyricsOptions || {})
    }
  );

  // Import the BGM API
  const backgroundMusic = (await import('./bgm.js')).default;
  
  // Build genre tags based on song style and mood
  const songStyle = foodData.songStyle || 'pop';
  const mood = foodData.mood || 'upbeat';
  let genreTags = `${mood} ${songStyle}`;
  
  // Add some common tags based on style
  if (songStyle.includes('pop')) {
    genreTags += ' vocal bright';
  } else if (songStyle.includes('rock')) {
    genreTags += ' guitar energetic';
  } else if (songStyle.includes('electronic')) {
    genreTags += ' electronic synthesizer';
  } else if (songStyle.includes('jazz')) {
    genreTags += ' saxophone smooth';
  } else {
    genreTags += ' vocal melodic';
  }

  // If yueConfig.outputDir not provided, use a default location
  const outputDir = apiConfig.yueConfig?.outputDir || './output/music';

  // Generate music using the lyrics and YuE
  const musicResult = await backgroundMusic.generate({
    genre: genreTags,
    lyrics,
    outputDir,
    stage1Model: apiConfig.yueConfig?.stage1Model,
    stage2Model: apiConfig.yueConfig?.stage2Model,
    referenceAudio: foodData.musicOptions?.referenceAudio
  });

  return {
    lyrics,
    musicResult
  };
}

export default {
  generateFoodVideoNarration,
  createFoodVideoWithNarration,
  generateFoodSongLyrics,
  createFoodSongWithLyrics
};
