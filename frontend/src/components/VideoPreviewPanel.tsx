import {
  Box,
  Card,
  CardBody,
  CardHeader,
  Heading,
  Text,
  Progress,
  Center,
  Icon,
  Spinner,
  useColorModeValue
} from '@chakra-ui/react'
import { FaVideo } from 'react-icons/fa'
import { ProgressData } from '../types'

interface VideoPreviewPanelProps {
  videoUrl: string | null
  isGenerating: boolean
  progress: ProgressData
}

export default function VideoPreviewPanel({
  videoUrl,
  isGenerating,
  progress
}: VideoPreviewPanelProps) {
  const cardBg = useColorModeValue('white', 'gray.700')
  const placeholderBg = useColorModeValue('gray.100', 'gray.600')
  const textColor = useColorModeValue('gray.500', 'gray.400')

  return (
    <Card shadow="md" borderRadius="lg" bg={cardBg} height="100%">
      <CardHeader>
        <Heading size="md">Video Preview</Heading>
      </CardHeader>
      <CardBody>
        {videoUrl ? (
          <Box borderRadius="md" overflow="hidden">
            <video 
              controls
              autoPlay
              loop
              src={videoUrl}
              style={{ width: '100%', borderRadius: '8px' }}
            />
          </Box>
        ) : (
          <Box 
            height="300px" 
            bg={placeholderBg} 
            borderRadius="md"
            display="flex"
            alignItems="center"
            justifyContent="center"
            flexDirection="column"
            p={6}
          >
            {isGenerating ? (
              <>
                <Spinner size="xl" mb={6} color="brand.500" />
                <Heading size="sm" mb={4} textAlign="center">
                  Generating Video...
                </Heading>
                <Progress
                  value={progress.percent}
                  colorScheme="brand"
                  hasStripe
                  isAnimated
                  size="sm"
                  width="100%"
                  borderRadius="full"
                  mb={3}
                />
                <Text fontSize="sm" textAlign="center" mb={2}>
                  {progress.message}
                </Text>
                <Text fontSize="xs" color={textColor} textAlign="center">
                  Current stage: {progress.stage}
                </Text>
              </>
            ) : (
              <>
                <Icon as={FaVideo} boxSize={12} color={textColor} mb={4} />
                <Heading size="sm" textAlign="center" mb={2}>
                  No video generated yet
                </Heading>
                <Text textAlign="center" color={textColor}>
                  Configure your video on the left and click "Generate Video"
                </Text>
              </>
            )}
          </Box>
        )}
      </CardBody>
    </Card>
  )
}
