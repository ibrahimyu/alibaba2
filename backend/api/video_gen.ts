class AlibabaAPI {
    private apiKey: string;
    private baseUrl: string = 'https://dashscope-intl.aliyuncs.com/api/v1/services/aigc/video-generation/video-synthesis';
    private tasksUrl: string = 'https://dashscope-intl.aliyuncs.com/api/v1/tasks';
    
    constructor(apiKey: string) {
        this.apiKey = apiKey;
    }

    /**
     * Generate video from an image using Alibaba Cloud's DashScope API
     * 
     * @param params.prompt - Text description of the video to generate
     * @param params.imageInput - Image input provided in one of the supported formats
     * @param params.imageType - Type of image input: 'url', 'base64', or 'formData'
     * @param params.resolution - Resolution of the output video (default: '720P')
     * @param params.promptExtend - Whether to extend the prompt (default: true)
     * @returns Promise with the API response containing task_id for checking status
     */
    async generateVideoFromImage({
        prompt,
        imageInput,
        imageType = 'url',
        resolution = '720P',
        promptExtend = true,
    }: {
        prompt: string;
        imageInput: string | FormData;
        imageType: 'url' | 'base64' | 'formData';
        resolution?: string;
        promptExtend?: boolean;
    }): Promise<VideoGenerationResponse> {
        try {
            const headers: Record<string, string> = {
                'X-DashScope-Async': 'enable',
                'Authorization': `Bearer ${this.apiKey}`,
            };
            
            let body: any;
            
            if (imageType === 'formData') {
                // For multipart/form-data, the FormData object is passed directly
                if (!(imageInput instanceof FormData)) {
                    throw new Error('When imageType is formData, imageInput must be a FormData instance');
                }
                
                // The FormData object should already contain all necessary fields
                return await this.sendFormDataRequest(imageInput, headers);
            } else {
                // For URL and base64 inputs, we use JSON body
                headers['Content-Type'] = 'application/json';
                
                body = {
                    model: 'wan2.1-i2v-turbo',
                    input: {
                        prompt,
                    },
                    parameters: {
                        resolution,
                        prompt_extend: promptExtend
                    }
                };
                
                // Add image input based on type
                if (imageType === 'url') {
                    body.input.img_url = imageInput;
                } else if (imageType === 'base64') {
                    body.input.img_base64 = imageInput;
                }
                
                return await this.sendJsonRequest(body, headers);
            }
        } catch (error) {
            console.error('Error generating video from image:', error);
            throw error;
        }
    }
    
    /**
     * Generate video from multiple reference images using Alibaba Cloud's DashScope API
     * 
     * @param params.prompt - Detailed description of the video to generate
     * @param params.refImagesUrls - Array of URLs for reference images
     * @param params.objOrBg - Array specifying if images are object or background references (default: ['obj', 'bg'])
     * @param params.size - Size of the output video (default: '1280*720')
     * @returns Promise with the API response containing task_id for checking status
     */
    async generateVideoFromMultipleImages({
        prompt,
        refImagesUrls,
        objOrBg = ['obj', 'bg'],
        size = '1280*720',
    }: {
        prompt: string;
        refImagesUrls: string[];
        objOrBg?: Array<'obj' | 'bg'>;
        size?: string;
    }): Promise<VideoGenerationResponse> {
        try {
            const headers: Record<string, string> = {
                'X-DashScope-Async': 'enable',
                'Authorization': `Bearer ${this.apiKey}`,
                'Content-Type': 'application/json',
            };
            
            const body = {
                model: 'wan2.1-vace-plus',
                input: {
                    function: 'image_reference',
                    prompt,
                    ref_images_url: refImagesUrls
                },
                parameters: {
                    obj_or_bg: objOrBg,
                    size
                }
            };
            
            return await this.sendJsonRequest(body, headers);
        } catch (error) {
            console.error('Error generating video from multiple images:', error);
            throw error;
        }
    }
    
    /**
     * Generate video by repainting an existing video using Alibaba Cloud's DashScope API
     * 
     * @param params.prompt - Detailed description of the video to generate
     * @param params.videoUrl - URL of the source video to repaint
     * @param params.controlCondition - Control condition for video repainting (default: 'depth')
     * @returns Promise with the API response containing task_id for checking status
     */
    async generateVideoRepainting({
        prompt,
        videoUrl,
        controlCondition = 'depth',
    }: {
        prompt: string;
        videoUrl: string;
        controlCondition?: 'depth' | string;
    }): Promise<VideoGenerationResponse> {
        try {
            const headers: Record<string, string> = {
                'X-DashScope-Async': 'enable',
                'Authorization': `Bearer ${this.apiKey}`,
                'Content-Type': 'application/json',
            };
            
            const body = {
                model: 'wan2.1-vace-plus',
                input: {
                    function: 'video_repainting',
                    prompt,
                    video_url: videoUrl
                },
                parameters: {
                    control_condition: controlCondition
                }
            };
            
            return await this.sendJsonRequest(body, headers);
        } catch (error) {
            console.error('Error generating video repainting:', error);
            throw error;
        }
    }
    
    /**
     * Check the status of a video generation task
     * 
     * @param taskId - The task ID returned from the initial generateVideoFromImage request
     * @returns Promise with the task status response
     */
    async checkTaskStatus(taskId: string): Promise<TaskStatusResponse> {
        try {
            const headers: Record<string, string> = {
                'Authorization': `Bearer ${this.apiKey}`,
            };
            
            const response = await fetch(`${this.tasksUrl}/${taskId}`, {
                method: 'GET',
                headers,
            });
            
            if (!response.ok) {
                throw new Error(`API request failed with status ${response.status}: ${await response.text()}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error checking task status:', error);
            throw error;
        }
    }
    
    /**
     * Poll for task completion with configurable intervals
     * 
     * @param taskId - The task ID to check
     * @param options.maxAttempts - Maximum number of polling attempts (default: 30)
     * @param options.intervalMs - Polling interval in milliseconds (default: 30000 - 30 seconds)
     * @returns Promise with the final task status response
     */
    async pollTaskCompletion(
        taskId: string, 
        { maxAttempts = 30, intervalMs = 30000 }: { maxAttempts?: number; intervalMs?: number } = {}
    ): Promise<TaskStatusResponse> {
        let attempts = 0;
        
        while (attempts < maxAttempts) {
            const response = await this.checkTaskStatus(taskId);
            
            if (response.output.task_status === 'SUCCEEDED' || 
                response.output.task_status === 'FAILED') {
                return response;
            }
            
            attempts++;
            
            // Wait for the specified interval
            await new Promise(resolve => setTimeout(resolve, intervalMs));
        }
        
        throw new Error(`Task polling timed out after ${maxAttempts} attempts`);
    }
    
    private async sendJsonRequest(body: any, headers: Record<string, string>): Promise<VideoGenerationResponse> {
        const response = await fetch(this.baseUrl, {
            method: 'POST',
            headers,
            body: JSON.stringify(body),
        });
        
        if (!response.ok) {
            throw new Error(`API request failed with status ${response.status}: ${await response.text()}`);
        }
        
        return await response.json();
    }
    
    private async sendFormDataRequest(formData: FormData, headers: Record<string, string>): Promise<VideoGenerationResponse> {
        // Don't set Content-Type as fetch will automatically set it with the boundary
        const response = await fetch(this.baseUrl, {
            method: 'POST',
            headers,
            body: formData,
        });
        
        if (!response.ok) {
            throw new Error(`API request failed with status ${response.status}: ${await response.text()}`);
        }
        
        return await response.json();
    }
}

/**
 * Response from the initial video generation API call
 */
interface VideoGenerationResponse {
    output: {
        task_status: 'PENDING' | 'RUNNING' | 'SUCCEEDED' | 'FAILED';
        task_id: string;
    };
    request_id: string;
    code?: string;
    message?: string;
}

/**
 * Response from checking task status
 */
interface TaskStatusResponse {
    request_id: string;
    output: {
        task_id: string;
        task_status: 'PENDING' | 'RUNNING' | 'SUCCEEDED' | 'FAILED';
        submit_time?: string;
        scheduled_time?: string;
        end_time?: string;
        video_url?: string;
        code?: string;
        message?: string;
    };
    usage?: {
        video_duration: number;
        video_ratio: string;
        video_count: number;
    };
}

export default AlibabaAPI;