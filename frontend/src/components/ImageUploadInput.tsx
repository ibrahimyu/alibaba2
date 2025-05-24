// ImageUploadInput.tsx
import {
  Box,
  Button,
  FormControl,
  FormErrorMessage,
  FormHelperText,
  FormLabel,
  Image,
  Input,
  InputGroup,
  InputRightElement,
  useToast
} from '@chakra-ui/react';
import { useRef, useState } from 'react';
import { FaCheck, FaUpload } from 'react-icons/fa';
import { uploadImage } from '../api';

interface ImageUploadInputProps {
  label: string;
  value: string;
  onChange: (url: string) => void;
  placeholder?: string;
  isRequired?: boolean;
  isInvalid?: boolean;
  errorMessage?: string;
}

export default function ImageUploadInput({ 
  label, 
  value, 
  onChange, 
  placeholder = 'Enter image URL or upload an image',
  isRequired = false,
  isInvalid = false,
  errorMessage = 'This field is required'
}: ImageUploadInputProps) {
  const [isUploading, setIsUploading] = useState(false);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const toast = useToast();

  const handleUploadClick = () => {
    if (fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Show local preview
    const localPreview = URL.createObjectURL(file);
    setPreviewUrl(localPreview);
    
    // Upload to server
    setIsUploading(true);
    
    try {
      // Use the uploadImage utility from api.ts
      const imageUrl = await uploadImage(file);
      onChange(imageUrl);
      
      toast({
        title: 'Image uploaded successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error) {
      console.error('Error uploading image:', error);
      toast({
        title: 'Upload failed',
        description: error instanceof Error ? error.message : 'Unknown error occurred',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      
      // Reset preview if upload fails
      setPreviewUrl(null);
    } finally {
      setIsUploading(false);
      // Clean up object URL to prevent memory leaks
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
    }
    
    // Clear the file input
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const imageUrl = value || previewUrl;
  
  return (
    <FormControl isInvalid={isInvalid} isRequired={isRequired}>
      <FormLabel>{label}</FormLabel>
      <InputGroup>
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          pr="4.5rem"
        />
        <InputRightElement width="4.5rem">
          <Button 
            h="1.75rem" 
            size="sm" 
            colorScheme="brand"
            onClick={handleUploadClick}
            isLoading={isUploading}
            leftIcon={value ? <FaCheck /> : <FaUpload />}
          >
            {isUploading ? 'Uploading' : value ? 'OK' : 'Upload'}
          </Button>
        </InputRightElement>
      </InputGroup>
      <input
        type="file"
        ref={fileInputRef}
        onChange={handleFileChange}
        accept="image/*"
        style={{ display: 'none' }}
      />
      {isInvalid && <FormErrorMessage>{errorMessage}</FormErrorMessage>}
      {imageUrl && (
        <Box mt={2}>
          <Image
            src={imageUrl}
            alt="Selected image"
            maxH="150px"
            borderRadius="md"
            fallbackSrc="https://via.placeholder.com/150x84?text=Loading+Image..."
            onError={() => {
              // Handle image load error
              toast({
                title: 'Error loading image',
                description: 'Unable to load the image from the provided URL',
                status: 'warning',
                duration: 3000,
                isClosable: true,
              });
            }}
          />
          <FormHelperText>Image will be resized to 1280x720 for the video</FormHelperText>
        </Box>
      )}
    </FormControl>
  );
}
