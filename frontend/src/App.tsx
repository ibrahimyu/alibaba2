import { Box, Container, Heading, useToast } from '@chakra-ui/react'
import { useState } from 'react'
import { VideoFormData } from './types'
import FormPanel from './components/FormPanel'
import VideoPreviewPanel from './components/VideoPreviewPanel'
import axios from 'axios'

// Default form data matches the sample_input_new.json structure
const defaultFormData: VideoFormData = {
  resto_name: "My Restaurant",
  resto_address: "123 Main Street",
  opening_scene: {
    prompt: "A beautiful restaurant with modern decor and ambient lighting",
    image_url: "https://example.com/opening_scene.jpg"
  },
  closing_scene: {
    prompt: "Customers leaving the restaurant with satisfied smiles",
    image_url: "https://example.com/closing_scene.jpg"
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
      photo_url: "https://example.com/dish.jpg"
    }
  ]
}

function App() {
  const [formData, setFormData] = useState<VideoFormData>(defaultFormData)
  const [isGenerating, setIsGenerating] = useState(false)
  const [jobId, setJobId] = useState<string | null>(null)
  const [progress, setProgress] = useState({ stage: '', percent: 0, message: '' })
  const [videoUrl, setVideoUrl] = useState<string | null>(null)
  const toast = useToast()

  const handleFormChange = (newData: VideoFormData) => {
    setFormData(newData)
  }

  const handleGenerate = async () => {
    try {
      setIsGenerating(true)
      setVideoUrl(null)
      
      // Submit the form data to the backend API
      const response = await axios.post('/api/generate-video', formData)
      const { jobId } = response.data
      
      if (jobId) {
        setJobId(jobId)
        
        // Start polling for progress
        const progressInterval = setInterval(async () => {
          try {
            const progressResponse = await axios.get(`/api/progress/${jobId}`)
            const jobData = progressResponse.data
            
            setProgress({
              stage: jobData.stage,
              percent: jobData.progress || 0,
              message: jobData.message
            })
            
            // If completed or failed, stop polling
            if (jobData.stage === 'completed') {
              clearInterval(progressInterval)
              setIsGenerating(false)
              
              // Extract video URL from the message
              const match = jobData.message.match(/available at (.+)$/i)
              if (match && match[1]) {
                setVideoUrl(match[1])
                toast({
                  title: 'Video generation complete!',
                  status: 'success',
                  duration: 5000,
                  isClosable: true
                })
              }
            } else if (jobData.stage === 'failed') {
              clearInterval(progressInterval)
              setIsGenerating(false)
              toast({
                title: 'Video generation failed',
                description: jobData.message,
                status: 'error',
                duration: 9000,
                isClosable: true
              })
            }
          } catch (error) {
            console.error('Error fetching progress:', error)
          }
        }, 2000) // Poll every 2 seconds
        
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
        GoBiz Video Generator
      </Heading>
      
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
    </Container>
  )
}

export default App
