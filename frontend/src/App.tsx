import { Box, Container, Heading, useToast, Tabs, TabList, TabPanels, Tab, TabPanel } from '@chakra-ui/react'
import { useState, useEffect, useCallback } from 'react'
import { VideoFormData, JobData } from './types'
import FormPanel from './components/FormPanel'
import VideoPreviewPanel from './components/VideoPreviewPanel'
import JobsPanel from './components/JobsPanel'
import { generateVideo, getJobProgress, getAllJobs } from './api'

const defaultFormData: VideoFormData = {
  resto_name: "My Restaurant",
  resto_address: "123 Main Street",
  opening_scene: {
    prompt: "A beautiful restaurant with modern decor and ambient lighting",
    image_url: ""
  },
  closing_scene: {
    prompt: "Customers leaving the restaurant with satisfied smiles",
    image_url: ""
  },
  music: {
    enabled: true,
    genres: "ambient lounge relaxing instrumental",
    bpm: 110,
    lyrics: "Instrumental music with a relaxing vibe"
  },
  menu: [
    {
      name: "Signature Dish",
      price: 25000,
      description: "Our chef's special creation",
      photo_url: ""
    }
  ]
}

function App() {
  const [formData, setFormData] = useState<VideoFormData>(defaultFormData)
  const [isGenerating, setIsGenerating] = useState(false)
  const [jobId, setJobId] = useState<string | null>(null)
  const [progress, setProgress] = useState({ stage: '', percent: 0, message: '', status: 'processing' as const })
  const [videoUrl, setVideoUrl] = useState<string | null>(null)
  const [jobs, setJobs] = useState<JobData[]>([])
  const [isLoadingJobs, setIsLoadingJobs] = useState(false)
  const [tabIndex, setTabIndex] = useState(0)
  const toast = useToast()

  // Function to fetch all jobs
  const fetchJobs = useCallback(async () => {
    try {
      setIsLoadingJobs(true);
      const jobsData = await getAllJobs();
      setJobs(jobsData);
    } catch (error) {
      console.error('Error fetching jobs:', error);
      toast({
        title: 'Error fetching jobs',
        description: error instanceof Error ? error.message : 'Unknown error',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setIsLoadingJobs(false);
    }
  }, [toast]);

  // Load jobs on component mount
  useEffect(() => {
    fetchJobs();
  }, [fetchJobs]);

  // Handle job selection from JobsPanel
  const handleJobSelected = async (jobId: string, videoUrl: string | null) => {
    setTabIndex(0); // Switch to main tab
    setJobId(jobId);
    
    if (videoUrl) {
      setVideoUrl(videoUrl);
      setIsGenerating(false);
    } else {
      setIsGenerating(true);
      setVideoUrl(null);
      
      // Start polling for progress
      startPollingJob(jobId);
    }
  };
  
  // Start polling a specific job
  const startPollingJob = (jobId: string) => {
    const progressInterval = setInterval(async () => {
      try {
        const jobData = await getJobProgress(jobId);
        
        setProgress({
          stage: jobData.stage,
          percent: jobData.percent || 0,
          message: jobData.message,
          status: jobData.status || 'processing'
        });
        
        // If completed or failed, stop polling
        if (jobData.status === 'completed') {
          clearInterval(progressInterval);
          setIsGenerating(false);
          setVideoUrl(jobData.video_url || null);
          
          toast({
            title: 'Video generation complete!',
            status: 'success',
            duration: 5000,
            isClosable: true
          });
          
          // Refresh job list
          fetchJobs();
        } else if (jobData.status === 'failed') {
          clearInterval(progressInterval);
          setIsGenerating(false);
          
          toast({
            title: 'Video generation failed',
            description: jobData.error || 'Unknown error',
            status: 'error',
            duration: 9000,
            isClosable: true
          });
          
          // Refresh job list
          fetchJobs();
        }
      } catch (error) {
        console.error('Error fetching progress:', error);
      }
    }, 2000);
    
    return progressInterval;
  };

  const handleFormChange = (newData: VideoFormData) => {
    setFormData(newData)
  }

  const handleGenerate = async () => {
    try {
      setIsGenerating(true)
      setVideoUrl(null)
      
      // Submit the form data to the backend API
      const response = await generateVideo(formData)
      const { jobId } = response
      
      if (jobId) {
        setJobId(jobId)
        
        // Start polling for progress
        const progressInterval = startPollingJob(jobId)
        
        // Clean up interval if component unmounts
        return () => clearInterval(progressInterval)
      }
    } catch (error) {
      console.error('Error generating video:', error)
      setIsGenerating(false)
      toast({
        title: 'Error',
        description: 'Failed to start video generation',
        status: 'error',
        duration: 9000,
        isClosable: true
      })
    }
  }

  return (
    <Container maxW="container.xl" py={6}>
      <Heading as="h1" mb={6} textAlign="center" size="xl" color="brand.700">
        GoBiz Showcase
      </Heading>
      
      <Tabs index={tabIndex} onChange={setTabIndex} variant="enclosed" mb={6}>
        <TabList>
          <Tab>Create Video</Tab>
          <Tab>Previous Jobs</Tab>
        </TabList>
        
        <TabPanels>
          <TabPanel p={0} pt={4}>
            {/* Create Video Tab */}
            <Box display="flex" flexDirection={{ base: 'column', lg: 'row' }} gap={6}>
              <Box flex="1" minW={{ base: 'full', lg: '500px' }}>
                <FormPanel 
                  formData={formData} 
                  onChange={handleFormChange} 
                  onGenerate={handleGenerate}
                  isGenerating={isGenerating}
                  progress={progress}
                />
              </Box>
              
              <Box flex="1">
                <VideoPreviewPanel 
                  videoUrl={videoUrl} 
                  isGenerating={isGenerating}
                  progress={progress}
                />
              </Box>
            </Box>
          </TabPanel>
          
          <TabPanel p={0} pt={4}>
            {/* Previous Jobs Tab */}
            <JobsPanel 
              jobs={jobs}
              isLoadingJobs={isLoadingJobs}
              onRefreshJobs={fetchJobs}
              onJobSelected={handleJobSelected}
              currentFormData={formData}
            />
          </TabPanel>
        </TabPanels>
      </Tabs>
    </Container>
  )
}

export default App
