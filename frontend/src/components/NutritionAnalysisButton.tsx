import { Button, Icon, Text, Tooltip, useToast } from '@chakra-ui/react';
import { useState } from 'react';
import { FaFlask } from 'react-icons/fa';
import { analyzeFoodImage } from '../api';
import { FoodAnalysisResult } from '../types';

interface NutritionAnalysisButtonProps {
  imageUrl: string;
  onAnalysisComplete: (result: FoodAnalysisResult) => void;
  isDisabled?: boolean;
}

export default function NutritionAnalysisButton({
  imageUrl,
  onAnalysisComplete,
  isDisabled = false
}: NutritionAnalysisButtonProps) {
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const toast = useToast();

  const handleAnalyzeClick = async () => {
    if (!imageUrl) {
      toast({
        title: "No image available",
        description: "Please upload a food image first.",
        status: "warning",
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    setIsAnalyzing(true);
    try {
      const result = await analyzeFoodImage(imageUrl);
      onAnalysisComplete(result);
      toast({
        title: "Analysis complete",
        description: "Nutritional information has been analyzed successfully.",
        status: "success",
        duration: 3000,
        isClosable: true,
      });
    } catch (error) {
      console.error("Analysis failed:", error);
      toast({
        title: "Analysis failed",
        description: error instanceof Error ? error.message : "Something went wrong.",
        status: "error",
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setIsAnalyzing(false);
    }
  };

  return (
    <Tooltip label="Analyze nutritional content of this food image">
      <Button
        size="sm"
        leftIcon={<Icon as={FaFlask} />}
        colorScheme="blue"
        variant="outline"
        onClick={handleAnalyzeClick}
        isLoading={isAnalyzing}
        loadingText="Analyzing"
        isDisabled={isDisabled || !imageUrl || isAnalyzing}
      >
        <Text>Analyze Nutrition</Text>
      </Button>
    </Tooltip>
  );
}
