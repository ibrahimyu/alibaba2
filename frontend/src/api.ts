// api.ts
import axios from 'axios';
import { VideoFormData } from './types';

/**
 * Upload an image file to the backend
 * The backend will resize it to 1280x720 and return a URL
 */
export const uploadImage = async (file: File): Promise<string> => {
  const formData = new FormData();
  formData.append('image', file);

  try {
    const response = await axios.post('/api/upload-image', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    if (response.data.success && response.data.url) {
      return response.data.url;
    } else {
      throw new Error(response.data.message || 'Upload failed');
    }
  } catch (error) {
    if (axios.isAxiosError(error) && error.response) {
      throw new Error(`Upload failed: ${error.response.data.message || error.message}`);
    }
    throw error;
  }
};

/**
 * Generate a video with the provided form data
 */
export const generateVideo = async (formData: VideoFormData) => {
  try {
    const response = await axios.post('/api/generate-video', formData);
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response) {
      throw new Error(`Video generation failed: ${error.response.data.message || error.message}`);
    }
    throw error;
  }
};

/**
 * Resume a failed video generation job
 */
export const resumeVideoGeneration = async (jobId: string, formData?: VideoFormData) => {
  try {
    const response = await axios.post(`/api/resume-video/${jobId}`, formData || {});
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response) {
      throw new Error(`Failed to resume generation: ${error.response.data.message || error.message}`);
    }
    throw error;
  }
};

/**
 * Fetch the progress of a video generation job
 */
export const getJobProgress = async (jobId: string) => {
  try {
    const response = await axios.get(`/api/progress/${jobId}`);
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response) {
      throw new Error(`Failed to fetch progress: ${error.response.data.message || error.message}`);
    }
    throw error;
  }
};

/**
 * Fetch all jobs
 */
export const getAllJobs = async () => {
  try {
    const response = await axios.get('/api/jobs');
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response) {
      throw new Error(`Failed to fetch jobs: ${error.response.data.message || error.message}`);
    }
    throw error;
  }
};

/**
 * Analyze a food image to get nutritional information
 */
export const analyzeFoodImage = async (imageUrl: string) => {
  try {
    const response = await axios.post('/api/analyze-food', { image_url: imageUrl });
    
    if (response.data.success && response.data.analysis) {
      return response.data.analysis;
    } else {
      throw new Error(response.data.message || 'Analysis failed');
    }
  } catch (error) {
    if (axios.isAxiosError(error) && error.response) {
      throw new Error(`Food analysis failed: ${error.response.data.message || error.message}`);
    }
    throw error;
  }
};
