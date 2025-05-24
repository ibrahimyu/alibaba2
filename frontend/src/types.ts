export interface MenuItem {
  name: string;
  price: number;
  description: string;
  photo_url: string;
}

export interface Scene {
  prompt: string;
  image_url: string;
}

export interface Music {
  enabled: boolean;
  genres: string;
  bpm?: number;
  lyrics?: string;
}

export interface VideoFormData {
  resto_name: string;
  resto_address: string;
  opening_scene: Scene;
  closing_scene: Scene;
  music: Music;
  menu: MenuItem[];
}

export interface ProgressData {
  stage: string;
  percent: number;
  message: string;
  status?: 'processing' | 'completed' | 'failed';
  error?: string;
}

export interface JobData {
  job_id: string;
  status: 'processing' | 'completed' | 'failed';
  stage: string;
  percent: number;
  message: string;
  video_url?: string;
  error?: string;
  start_time: string;
  update_time: string;
}
