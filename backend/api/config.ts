// filepath: /Users/ibrahim/Documents/alisandbox/api/config.ts

/**
 * Global configuration module that loads environment variables and provides
 * access to API keys and other configuration properties.
 */

// Load environment variables
import 'dotenv/config';
import AlibabaAPI from './video_gen.js';

/**
 * Configuration singleton that provides access to API keys and other configuration.
 */
class Config {
  // API keys
  public readonly openAiApiKey: string;
  public readonly alibabaApiKey: string;
  
  // API clients
  public readonly alibabaApi: AlibabaAPI;
  
  constructor() {
    // Load API keys from environment variables
    this.openAiApiKey = process.env.OPENAI_API_KEY || '';
    this.alibabaApiKey = process.env.ALIBABA_API_KEY || process.env.DASHSCOPE_API_KEY || '';
    
    // Validate API keys
    if (!this.openAiApiKey) {
      throw new Error('OpenAI API key not found. Please set OPENAI_API_KEY environment variable.');
    }
    
    if (!this.alibabaApiKey) {
      throw new Error('Alibaba API key not found. Please set ALIBABA_API_KEY or DASHSCOPE_API_KEY environment variable.');
    }
    
    // Initialize API clients
    this.alibabaApi = new AlibabaAPI(this.alibabaApiKey);
  }
}

// Create and export a singleton instance
const config = new Config();
export default config;