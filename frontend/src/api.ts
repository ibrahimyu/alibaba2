// api.ts
import axios from 'axios';

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
export const generateVideo = async (formData: any) => {
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
