import {
  Box,
  Button,
  Card,
  CardBody,
  CardHeader,
  FormControl,
  FormLabel,
  Heading,
  Input,
  Switch,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
  Text,
  Textarea,
  VStack,
  Progress,
  Flex,
  Icon,
  useColorModeValue,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  Divider,
  IconButton
} from '@chakra-ui/react'
import { useState } from 'react'
import { VideoFormData, MenuItem, ProgressData } from '../types'
import { FaPlay, FaPlus, FaTrash } from 'react-icons/fa'
import ImageUploadInput from './ImageUploadInput'

interface FormPanelProps {
  formData: VideoFormData
  onChange: (data: VideoFormData) => void
  onGenerate: () => void
  isGenerating: boolean
  progress: ProgressData
}

export default function FormPanel({
  formData,
  onChange,
  onGenerate,
  isGenerating,
  progress
}: FormPanelProps) {
  const [activeTab, setActiveTab] = useState(0)
  const cardBg = useColorModeValue('white', 'gray.700')
  const textColor = useColorModeValue('gray.600', 'gray.300')

  // Update form data with new values
  const updateFormData = (path: string, value: any) => {
    const newData = { ...formData }
    const parts = path.split('.')
    let current: any = newData
    
    for (let i = 0; i < parts.length - 1; i++) {
      const part = parts[i]
      if (part.includes('[')) {
        const [arrayName, indexStr] = part.split('[')
        const index = parseInt(indexStr.replace(']', ''))
        current = current[arrayName][index]
      } else {
        current = current[part]
      }
    }
    
    const lastPart = parts[parts.length - 1]
    current[lastPart] = value
    
    onChange(newData)
  }

  const addMenuItem = () => {
    const newData = { ...formData }
    newData.menu.push({
      name: `Menu Item ${newData.menu.length + 1}`,
      price: 10000,
      description: 'Description of the menu item',
      photo_url: ''
    })
    onChange(newData)
  }

  const removeMenuItem = (index: number) => {
    const newData = { ...formData }
    newData.menu.splice(index, 1)
    onChange(newData)
  }

  return (
    <Card shadow="md" borderRadius="lg" bg={cardBg} height="100%">
      <CardHeader pb={0}>
        <Heading size="md" mb={4}>Video Configuration</Heading>
        {isGenerating && (
          <Box mb={4}>
            <Text mb={2} fontWeight="medium">{progress.message}</Text>
            <Progress
              value={progress.percent}
              colorScheme="brand"
              hasStripe
              isAnimated
              size="sm"
              borderRadius="full"
            />
            <Text mt={1} fontSize="sm" color={textColor}>
              {progress.stage} - {progress.percent}%
            </Text>
          </Box>
        )}
      </CardHeader>
      <CardBody>
        <Tabs isLazy variant="enclosed" colorScheme="brand" index={activeTab} onChange={setActiveTab}>
          <TabList mb={4}>
            <Tab>Basic Info</Tab>
            <Tab>Opening Scene</Tab>
            <Tab>Menu Items</Tab>
            <Tab>Closing Scene</Tab>
            <Tab>Music</Tab>
          </TabList>
          
          <TabPanels>
            {/* Basic Info Tab */}
            <TabPanel>
              <VStack spacing={4} align="stretch">
                <FormControl>
                  <FormLabel>Restaurant Name</FormLabel>
                  <Input 
                    value={formData.resto_name}
                    onChange={(e) => updateFormData('resto_name', e.target.value)}
                    placeholder="Enter restaurant name"
                  />
                </FormControl>
                
                <FormControl>
                  <FormLabel>Restaurant Address</FormLabel>
                  <Input 
                    value={formData.resto_address}
                    onChange={(e) => updateFormData('resto_address', e.target.value)}
                    placeholder="Enter restaurant address"
                  />
                </FormControl>
              </VStack>
            </TabPanel>
            
            {/* Opening Scene Tab */}
            <TabPanel>
              <VStack spacing={4} align="stretch">
                <FormControl>
                  <FormLabel>Opening Scene Description</FormLabel>
                  <Textarea 
                    value={formData.opening_scene.prompt}
                    onChange={(e) => updateFormData('opening_scene.prompt', e.target.value)}
                    placeholder="Describe the opening scene of your video"
                    minH="100px"
                  />
                </FormControl>
                
                <ImageUploadInput
                  label="Opening Scene Image"
                  value={formData.opening_scene.image_url}
                  onChange={(url) => updateFormData('opening_scene.image_url', url)}
                  placeholder="Enter image URL or upload an image"
                />
              </VStack>
            </TabPanel>
            
            {/* Menu Items Tab */}
            <TabPanel>
              <VStack spacing={6} align="stretch">
                {formData.menu.map((item: MenuItem, index: number) => (
                  <Box key={index} p={4} borderWidth={1} borderRadius="md">
                    <Flex justify="space-between" mb={2}>
                      <Heading size="sm">Menu Item {index + 1}</Heading>
                      <IconButton
                        aria-label="Remove menu item"
                        icon={<FaTrash />}
                        size="sm"
                        colorScheme="red"
                        variant="ghost"
                        onClick={() => removeMenuItem(index)}
                      />
                    </Flex>
                    <Divider mb={4} />
                    
                    <VStack spacing={3} align="stretch">
                      <FormControl>
                        <FormLabel>Name</FormLabel>
                        <Input 
                          value={item.name}
                          onChange={(e) => updateFormData(`menu[${index}].name`, e.target.value)}
                          placeholder="Enter item name"
                        />
                      </FormControl>
                      
                      <FormControl>
                        <FormLabel>Price</FormLabel>
                        <NumberInput
                          value={item.price}
                          onChange={(valueStr) => updateFormData(`menu[${index}].price`, parseInt(valueStr))}
                          min={0}
                        >
                          <NumberInputField />
                          <NumberInputStepper>
                            <NumberIncrementStepper />
                            <NumberDecrementStepper />
                          </NumberInputStepper>
                        </NumberInput>
                      </FormControl>
                      
                      <FormControl>
                        <FormLabel>Description</FormLabel>
                        <Textarea 
                          value={item.description}
                          onChange={(e) => updateFormData(`menu[${index}].description`, e.target.value)}
                          placeholder="Describe this menu item"
                        />
                      </FormControl>
                      
                      <ImageUploadInput
                        label="Food Photo"
                        value={item.photo_url}
                        onChange={(url) => updateFormData(`menu[${index}].photo_url`, url)}
                        placeholder="Enter food photo URL or upload an image"
                      />
                    </VStack>
                  </Box>
                ))}
                
                <Button 
                  leftIcon={<FaPlus />} 
                  colorScheme="brand" 
                  variant="outline"
                  onClick={addMenuItem}
                >
                  Add Menu Item
                </Button>
              </VStack>
            </TabPanel>
            
            {/* Closing Scene Tab */}
            <TabPanel>
              <VStack spacing={4} align="stretch">
                <FormControl>
                  <FormLabel>Closing Scene Description</FormLabel>
                  <Textarea 
                    value={formData.closing_scene.prompt}
                    onChange={(e) => updateFormData('closing_scene.prompt', e.target.value)}
                    placeholder="Describe the closing scene of your video"
                    minH="100px"
                  />
                </FormControl>
                
                <ImageUploadInput
                  label="Closing Scene Image"
                  value={formData.closing_scene.image_url}
                  onChange={(url) => updateFormData('closing_scene.image_url', url)}
                  placeholder="Enter image URL or upload an image"
                />
              </VStack>
            </TabPanel>
            
            {/* Music Tab */}
            <TabPanel>
              <VStack spacing={4} align="stretch">
                <FormControl display="flex" alignItems="center">
                  <FormLabel mb={0}>Enable Background Music</FormLabel>
                  <Switch 
                    isChecked={formData.music.enabled}
                    onChange={(e) => updateFormData('music.enabled', e.target.checked)}
                    colorScheme="brand"
                  />
                </FormControl>
                
                {formData.music.enabled && (
                  <>
                    <FormControl>
                      <FormLabel>Music Genres</FormLabel>
                      <Input 
                        value={formData.music.genres}
                        onChange={(e) => updateFormData('music.genres', e.target.value)}
                        placeholder="e.g., ambient lounge relaxing instrumental"
                      />
                    </FormControl>
                    
                    <FormControl>
                      <FormLabel>BPM (Beats Per Minute)</FormLabel>
                      <NumberInput
                        value={formData.music.bpm || 120}
                        onChange={(valueStr) => updateFormData('music.bpm', parseInt(valueStr))}
                        min={60}
                        max={200}
                      >
                        <NumberInputField />
                        <NumberInputStepper>
                          <NumberIncrementStepper />
                          <NumberDecrementStepper />
                        </NumberInputStepper>
                      </NumberInput>
                    </FormControl>
                    
                    <FormControl>
                      <FormLabel>Lyrics (Optional)</FormLabel>
                      <Textarea 
                        value={formData.music.lyrics || ''}
                        onChange={(e) => updateFormData('music.lyrics', e.target.value)}
                        placeholder="Enter lyrics or leave blank for instrumental"
                      />
                    </FormControl>
                  </>
                )}
              </VStack>
            </TabPanel>
          </TabPanels>
        </Tabs>
        
        <Flex justifyContent="flex-end" mt={6}>
          <Button
            leftIcon={<Icon as={FaPlay} />}
            colorScheme="brand"
            size="lg"
            onClick={onGenerate}
            isLoading={isGenerating}
            loadingText={isGenerating ? progress.message : undefined}
            isDisabled={isGenerating}
          >
            Generate Video
          </Button>
        </Flex>
      </CardBody>
    </Card>
  )
}
