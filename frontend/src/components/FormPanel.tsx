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
  IconButton,
  Tooltip
} from '@chakra-ui/react'
import { useState, useMemo } from 'react'
import { VideoFormData, MenuItem, ProgressData, FoodAnalysisResult } from '../types'
import { FaPlay, FaPlus, FaTrash } from 'react-icons/fa'
import { BsExclamationCircleFill } from 'react-icons/bs'
import ImageUploadInput from './ImageUploadInput'
import NutritionAnalysisButton from './NutritionAnalysisButton'

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
  const [validationErrors, setValidationErrors] = useState<Record<string, string[]>>({})
  
  // Validate the form data
  const isFormValid = useMemo(() => {
    const errors: Record<string, string[]> = {}
    
    // Check restaurant info
    if (!formData.resto_name || formData.resto_name.trim() === '') {
      if (!errors.basicInfo) errors.basicInfo = []
      errors.basicInfo.push('Restaurant name is required')
    }
    
    if (!formData.resto_address || formData.resto_address.trim() === '') {
      if (!errors.basicInfo) errors.basicInfo = []
      errors.basicInfo.push('Restaurant address is required')
    }
    
    // Check opening scene
    if (!formData.opening_scene.prompt || formData.opening_scene.prompt.trim() === '') {
      if (!errors.openingScene) errors.openingScene = []
      errors.openingScene.push('Opening scene description is required')
    }
    
    if (!formData.opening_scene.image_url || formData.opening_scene.image_url.trim() === '') {
      if (!errors.openingScene) errors.openingScene = []
      errors.openingScene.push('Opening scene image is required')
    }
    
    // Check closing scene
    if (!formData.closing_scene.prompt || formData.closing_scene.prompt.trim() === '') {
      if (!errors.closingScene) errors.closingScene = []
      errors.closingScene.push('Closing scene description is required')
    }
    
    if (!formData.closing_scene.image_url || formData.closing_scene.image_url.trim() === '') {
      if (!errors.closingScene) errors.closingScene = []
      errors.closingScene.push('Closing scene image is required')
    }
    
    // Check menu items
    formData.menu.forEach((item, index) => {
      const errorKey = `menuItem-${index}`
      if (!item.name || item.name.trim() === '') {
        if (!errors[errorKey]) errors[errorKey] = []
        errors[errorKey].push('Menu item name is required')
      }
      
      if (!item.description || item.description.trim() === '') {
        if (!errors[errorKey]) errors[errorKey] = []
        errors[errorKey].push('Menu item description is required')
      }
      
      if (!item.photo_url || item.photo_url.trim() === '') {
        if (!errors[errorKey]) errors[errorKey] = []
        errors[errorKey].push('Menu item image is required')
      }
    })
    
    // If music is enabled, check required music fields
    if (formData.music.enabled) {
      if (!formData.music.genres || formData.music.genres.trim() === '') {
        if (!errors.music) errors.music = []
        errors.music.push('Music genres are required when music is enabled')
      }
    }
    
    // Save the errors for display
    setValidationErrors(errors)
    
    // Form is valid if there are no errors
    return Object.keys(errors).length === 0
  }, [formData])

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

  // Helper function to check if a tab has errors
  const hasTabErrors = (tabName: string) => {
    return Object.keys(validationErrors).some(key => key.startsWith(tabName))
  }

  // NutritionDisplay component to show food nutrition information
  function NutritionDisplay({ nutrition }: { nutrition?: FoodAnalysisResult }) {
    if (!nutrition || !nutrition.total_nutrition) return null;
    
    const total = nutrition.total_nutrition;
    
    return (
      <Box mt={3} p={3} borderWidth={1} borderRadius="md" bg="gray.50" _dark={{ bg: "gray.700" }}>
        <Heading size="xs" mb={2}>Nutritional Information</Heading>
        <Flex flexWrap="wrap">
          <Box flex="1" minW="120px" p={1}>
            <Text fontWeight="bold" fontSize="sm">Calories:</Text>
            <Text fontSize="sm">{total.calories}</Text>
          </Box>
          <Box flex="1" minW="120px" p={1}>
            <Text fontWeight="bold" fontSize="sm">Protein:</Text>
            <Text fontSize="sm">{total.protein}</Text>
          </Box>
          <Box flex="1" minW="120px" p={1}>
            <Text fontWeight="bold" fontSize="sm">Carbs:</Text>
            <Text fontSize="sm">{total.carbs}</Text>
          </Box>
          <Box flex="1" minW="120px" p={1}>
            <Text fontWeight="bold" fontSize="sm">Fat:</Text>
            <Text fontSize="sm">{total.fat}</Text>
          </Box>
          {total.fiber && (
            <Box flex="1" minW="120px" p={1}>
              <Text fontWeight="bold" fontSize="sm">Fiber:</Text>
              <Text fontSize="sm">{total.fiber}</Text>
            </Box>
          )}
          {total.sodium && (
            <Box flex="1" minW="120px" p={1}>
              <Text fontWeight="bold" fontSize="sm">Sodium:</Text>
              <Text fontSize="sm">{total.sodium}</Text>
            </Box>
          )}
        </Flex>
        {nutrition.foods && nutrition.foods.length > 1 && (
          <Tooltip label="Full nutrition breakdown available" hasArrow>
            <Text fontSize="xs" color="blue.500" mt={1} cursor="pointer">
              Contains data for {nutrition.foods.length} food items
            </Text>
          </Tooltip>
        )}
        {nutrition.foods && nutrition.foods.length > 0 && (
          <Box mt={2}>
            <Divider my={2} />
            <Heading size="xs" mb={2}>Individual Food Items</Heading>
            {nutrition.foods.map((food, idx) => (
              <Box key={idx} p={2} mt={1} borderWidth={1} borderRadius="sm" fontSize="xs">
                <Text fontWeight="medium">{food.name}</Text>
                <Flex flexWrap="wrap" mt={1}>
                  <Text mr={2}>Calories: {food.calories}</Text>
                  <Text mr={2}>Protein: {food.protein}</Text>
                  <Text mr={2}>Carbs: {food.carbs}</Text>
                  <Text mr={2}>Fat: {food.fat}</Text>
                  {food.fiber && <Text mr={2}>Fiber: {food.fiber}</Text>}
                  {food.sodium && <Text mr={2}>Sodium: {food.sodium}</Text>}
                </Flex>
              </Box>
            ))}
          </Box>
        )}
      </Box>
    );
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
            <Tab position="relative">
              Basic Info
              {hasTabErrors('basicInfo') && (
                <Icon as={BsExclamationCircleFill} color="red.500" ml={2} />
              )}
            </Tab>
            <Tab position="relative">
              Opening Scene
              {hasTabErrors('openingScene') && (
                <Icon as={BsExclamationCircleFill} color="red.500" ml={2} />
              )}
            </Tab>
            <Tab position="relative">
              Menu Items
              {Object.keys(validationErrors).some(key => key.startsWith('menuItem')) && (
                <Icon as={BsExclamationCircleFill} color="red.500" ml={2} />
              )}
            </Tab>
            <Tab position="relative">
              Closing Scene
              {hasTabErrors('closingScene') && (
                <Icon as={BsExclamationCircleFill} color="red.500" ml={2} />
              )}
            </Tab>
            <Tab position="relative">
              Music
              {hasTabErrors('music') && (
                <Icon as={BsExclamationCircleFill} color="red.500" ml={2} />
              )}
            </Tab>
          </TabList>
          
          <TabPanels>
            {/* Basic Info Tab */}
            <TabPanel>
              <VStack spacing={4} align="stretch">
                <FormControl isInvalid={!!validationErrors.basicInfo?.includes('Restaurant name is required')}>
                  <FormLabel>Restaurant Name</FormLabel>
                  <Input 
                    value={formData.resto_name}
                    onChange={(e) => updateFormData('resto_name', e.target.value)}
                    placeholder="Enter restaurant name"
                  />
                  {validationErrors.basicInfo?.includes('Restaurant name is required') && (
                    <Text color="red.500" fontSize="sm" mt={1}>Restaurant name is required</Text>
                  )}
                </FormControl>
                
                <FormControl isInvalid={!!validationErrors.basicInfo?.includes('Restaurant address is required')}>
                  <FormLabel>Restaurant Address</FormLabel>
                  <Input 
                    value={formData.resto_address}
                    onChange={(e) => updateFormData('resto_address', e.target.value)}
                    placeholder="Enter restaurant address"
                  />
                  {validationErrors.basicInfo?.includes('Restaurant address is required') && (
                    <Text color="red.500" fontSize="sm" mt={1}>Restaurant address is required</Text>
                  )}
                </FormControl>
              </VStack>
            </TabPanel>
            
            {/* Opening Scene Tab */}
            <TabPanel>
              <VStack spacing={4} align="stretch">
                <FormControl isInvalid={!!validationErrors.openingScene?.includes('Opening scene description is required')}>
                  <FormLabel>Opening Scene Description</FormLabel>
                  <Textarea 
                    value={formData.opening_scene.prompt}
                    onChange={(e) => updateFormData('opening_scene.prompt', e.target.value)}
                    placeholder="Describe the opening scene of your video"
                    minH="100px"
                  />
                  {validationErrors.openingScene?.includes('Opening scene description is required') && (
                    <Text color="red.500" fontSize="sm" mt={1}>Opening scene description is required</Text>
                  )}
                </FormControl>
                
                <ImageUploadInput
                  label="Opening Scene Image"
                  value={formData.opening_scene.image_url}
                  onChange={(url) => updateFormData('opening_scene.image_url', url)}
                  placeholder="Enter image URL or upload an image"
                  isRequired={true}
                  isInvalid={!!validationErrors.openingScene?.includes('Opening scene image is required')}
                  errorMessage="Opening scene image is required"
                />
              </VStack>
            </TabPanel>
            
            {/* Menu Items Tab */}
            <TabPanel>
              <VStack spacing={6} align="stretch">
                {formData.menu.map((item: MenuItem, index: number) => {
                  const errorKey = `menuItem-${index}`;
                  return (
                    <Box key={index} p={4} borderWidth={1} borderRadius="md" borderColor={validationErrors[errorKey] ? "red.300" : "inherit"}>
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
                        <FormControl isInvalid={!!validationErrors[errorKey]?.includes('Menu item name is required')}>
                          <FormLabel>Name</FormLabel>
                          <Input 
                            value={item.name}
                            onChange={(e) => updateFormData(`menu[${index}].name`, e.target.value)}
                            placeholder="Enter item name"
                          />
                          {validationErrors[errorKey]?.includes('Menu item name is required') && (
                            <Text color="red.500" fontSize="sm" mt={1}>Menu item name is required</Text>
                          )}
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
                        
                        <FormControl isInvalid={!!validationErrors[errorKey]?.includes('Menu item description is required')}>
                          <FormLabel>Description</FormLabel>
                          <Textarea 
                            value={item.description}
                            onChange={(e) => updateFormData(`menu[${index}].description`, e.target.value)}
                            placeholder="Describe this menu item"
                          />
                          {validationErrors[errorKey]?.includes('Menu item description is required') && (
                            <Text color="red.500" fontSize="sm" mt={1}>Menu item description is required</Text>
                          )}
                        </FormControl>
                        
                        <ImageUploadInput
                          label="Food Photo"
                          value={item.photo_url}
                          onChange={(url) => updateFormData(`menu[${index}].photo_url`, url)}
                          placeholder="Enter food photo URL or upload an image"
                          isRequired={true}
                          isInvalid={!!validationErrors[errorKey]?.includes('Menu item image is required')}
                          errorMessage="Menu item image is required"
                        />
                        
                        <Flex mt={2} justifyContent="flex-end">
                          <NutritionAnalysisButton 
                            imageUrl={item.photo_url}
                            onAnalysisComplete={(result) => updateFormData(`menu[${index}].nutrition`, result)}
                            isDisabled={!item.photo_url}
                          />
                        </Flex>
                        
                        {/* Show nutrition information if available */}
                        <NutritionDisplay nutrition={item.nutrition} />
                        
                        {/* Nutrition Information Display */}
                        <NutritionDisplay nutrition={item.nutrition} />
                      </VStack>
                    </Box>
                  );
                })}
                
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
                <FormControl isInvalid={!!validationErrors.closingScene?.includes('Closing scene description is required')}>
                  <FormLabel>Closing Scene Description</FormLabel>
                  <Textarea 
                    value={formData.closing_scene.prompt}
                    onChange={(e) => updateFormData('closing_scene.prompt', e.target.value)}
                    placeholder="Describe the closing scene of your video"
                    minH="100px"
                  />
                  {validationErrors.closingScene?.includes('Closing scene description is required') && (
                    <Text color="red.500" fontSize="sm" mt={1}>Closing scene description is required</Text>
                  )}
                </FormControl>
                
                <ImageUploadInput
                  label="Closing Scene Image"
                  value={formData.closing_scene.image_url}
                  onChange={(url) => updateFormData('closing_scene.image_url', url)}
                  placeholder="Enter image URL or upload an image"
                  isRequired={true}
                  isInvalid={!!validationErrors.closingScene?.includes('Closing scene image is required')}
                  errorMessage="Closing scene image is required"
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
                    <FormControl isInvalid={!!validationErrors.music?.includes('Music genres are required when music is enabled')}>
                      <FormLabel>Music Genres</FormLabel>
                      <Input 
                        value={formData.music.genres}
                        onChange={(e) => updateFormData('music.genres', e.target.value)}
                        placeholder="e.g., ambient lounge relaxing instrumental"
                      />
                      {validationErrors.music?.includes('Music genres are required when music is enabled') && (
                        <Text color="red.500" fontSize="sm" mt={1}>Music genres are required</Text>
                      )}
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
          <Tooltip 
            isDisabled={isFormValid}
            hasArrow
            label="Please fill in all required fields before generating the video"
            placement="top"
          >
            <Button
              leftIcon={<Icon as={FaPlay} />}
              colorScheme="brand"
              size="lg"
              onClick={onGenerate}
              isLoading={isGenerating}
              loadingText={isGenerating ? progress.message : undefined}
              isDisabled={isGenerating || !isFormValid}
            >
              Generate Video
            </Button>
          </Tooltip>
        </Flex>
      </CardBody>
    </Card>
  )
}
