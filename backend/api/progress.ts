// Progress tracker module for video generation
import EventEmitter from 'events';

class ProgressTracker extends EventEmitter {
  private static instance: ProgressTracker;
  private progressMap: Map<string, {
    stage: string;
    progress: number;
    message: string;
    started: Date;
    completed?: Date;
  }>;

  private constructor() {
    super();
    this.progressMap = new Map();
  }

  public static getInstance(): ProgressTracker {
    if (!ProgressTracker.instance) {
      ProgressTracker.instance = new ProgressTracker();
    }
    return ProgressTracker.instance;
  }

  /**
   * Start tracking progress for a new job
   */
  public startJob(jobId: string): void {
    this.progressMap.set(jobId, {
      stage: 'initializing',
      progress: 0,
      message: 'Starting video generation process',
      started: new Date()
    });
    this.emit('progress', this.getJobProgress(jobId));
  }

  /**
   * Update progress for an existing job
   */
  public updateProgress(jobId: string, stage: string, progress: number, message: string): void {
    if (!this.progressMap.has(jobId)) {
      this.startJob(jobId);
    }
    
    this.progressMap.set(jobId, {
      ...this.progressMap.get(jobId)!,
      stage,
      progress: Math.min(Math.max(progress, 0), 100),
      message
    });
    
    this.emit('progress', this.getJobProgress(jobId));
  }

  /**
   * Complete a job
   */
  public completeJob(jobId: string, message = 'Video generation completed'): void {
    if (!this.progressMap.has(jobId)) {
      this.startJob(jobId);
    }
    
    this.progressMap.set(jobId, {
      ...this.progressMap.get(jobId)!,
      stage: 'completed',
      progress: 100,
      message,
      completed: new Date()
    });
    
    this.emit('progress', this.getJobProgress(jobId));
    
    // Clean up after 1 hour
    setTimeout(() => {
      this.progressMap.delete(jobId);
    }, 60 * 60 * 1000);
  }

  /**
   * Mark a job as failed
   */
  public failJob(jobId: string, errorMessage: string): void {
    if (!this.progressMap.has(jobId)) {
      this.startJob(jobId);
    }
    
    this.progressMap.set(jobId, {
      ...this.progressMap.get(jobId)!,
      stage: 'failed',
      message: `Error: ${errorMessage}`,
      completed: new Date()
    });
    
    this.emit('progress', this.getJobProgress(jobId));
  }

  /**
   * Get the current progress of a job
   */
  public getJobProgress(jobId: string) {
    const progress = this.progressMap.get(jobId);
    if (!progress) {
      return {
        jobId,
        exists: false,
        progress: 0,
        stage: 'unknown',
        message: 'Job not found'
      };
    }
    
    return {
      jobId,
      exists: true,
      ...progress,
      elapsedMs: progress.completed ? 
        progress.completed.getTime() - progress.started.getTime() : 
        new Date().getTime() - progress.started.getTime()
    };
  }

  /**
   * Get all current jobs
   */
  public getAllJobs() {
    const result: any[] = [];
    this.progressMap.forEach((value, key) => {
      result.push({
        jobId: key,
        ...value,
        elapsedMs: value.completed ? 
          value.completed.getTime() - value.started.getTime() : 
          new Date().getTime() - value.started.getTime()
      });
    });
    return result;
  }
}

export default ProgressTracker.getInstance();
