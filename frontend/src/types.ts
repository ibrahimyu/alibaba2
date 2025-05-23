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
}
