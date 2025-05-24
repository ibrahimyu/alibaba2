import {
  Badge,
  Box,
  Button,
  Card,
  CardBody,
  CardHeader,
  Flex,
  Heading,
  Icon,
  Modal,
  ModalBody,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Spinner,
  Table,
  Tag,
  TagLabel,
  TagLeftIcon,
  Tbody,
  Td,
  Text,
  Th,
  Thead,
  Tooltip,
  Tr,
  useDisclosure,
  useToast
} from '@chakra-ui/react';
import { useState } from 'react';
import { FaCheck, FaExclamationTriangle, FaRedo } from 'react-icons/fa';
import { MdRefresh, MdVisibility } from 'react-icons/md';
import { resumeVideoGeneration } from '../api';
import { JobData, VideoFormData } from '../types';

interface JobsPanelProps {
  onJobSelected: (jobId: string, videoUrl: string | null) => void;
  currentFormData: VideoFormData;
  onRefreshJobs: () => void;
  isLoadingJobs: boolean;
  jobs: JobData[];
}

export default function JobsPanel({
  onJobSelected,
  currentFormData,
  onRefreshJobs,
  isLoadingJobs,
  jobs
}: JobsPanelProps) {
  const [isResuming, setIsResuming] = useState<Record<string, boolean>>({});
  const toast = useToast();
  const { isOpen, onOpen, onClose } = useDisclosure();
  const [selectedJob, setSelectedJob] = useState<JobData | null>(null);

  // Function to format date
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString();
  };
  
  // Function to get status color
  const getStatusColor = (status: string) => {
    switch(status) {
      case 'completed': return 'green';
      case 'failed': return 'red';
      case 'processing': return 'blue';
      default: return 'gray';
    }
  };

  // Handle resuming a job
  const handleResumeJob = async (jobId: string) => {
    try {
      setIsResuming(prev => ({ ...prev, [jobId]: true }));
      
      await resumeVideoGeneration(jobId, currentFormData);
      
      toast({
        title: 'Job Resumed',
        description: `Video generation for job ${jobId.slice(0, 8)}... has been resumed.`,
        status: 'success',
        duration: 5000,
        isClosable: true
      });
      
      // Refresh the jobs list
      onRefreshJobs();
    } catch (error) {
      toast({
        title: 'Resume Failed',
        description: error instanceof Error ? error.message : 'Unknown error occurred',
        status: 'error',
        duration: 5000,
        isClosable: true
      });
    } finally {
      setIsResuming(prev => ({ ...prev, [jobId]: false }));
    }
  };

  // Handle selecting a job
  const handleSelectJob = (job: JobData) => {
    setSelectedJob(job);
    onOpen();
  };

  // Handle confirming job selection
  const handleConfirmJobSelection = () => {
    if (selectedJob) {
      onJobSelected(selectedJob.job_id, selectedJob.video_url || null);
      onClose();
    }
  };

  return (
    <Card shadow="md" borderRadius="lg" height="100%" overflow="hidden">
      <CardHeader>
        <Flex justifyContent="space-between" alignItems="center">
          <Heading size="md">Previous Jobs</Heading>
          <Tooltip label="Refresh jobs list">
            <Button
              size="sm"
              leftIcon={<Icon as={MdRefresh} />}
              onClick={onRefreshJobs}
              isLoading={isLoadingJobs}
              loadingText="Refreshing"
              variant="outline"
            >
              Refresh
            </Button>
          </Tooltip>
        </Flex>
      </CardHeader>
      <CardBody overflow="auto" maxH="500px">
        {isLoadingJobs ? (
          <Flex justifyContent="center" py={6}>
            <Spinner />
          </Flex>
        ) : jobs.length === 0 ? (
          <Text textAlign="center" py={6} color="gray.500">
            No previous jobs found
          </Text>
        ) : (
          <Table variant="simple" size="sm">
            <Thead>
              <Tr>
                <Th>Status</Th>
                <Th>Started</Th>
                <Th>Message</Th>
                <Th>Actions</Th>
              </Tr>
            </Thead>
            <Tbody>
              {jobs.map((job) => (
                <Tr key={job.job_id}>
                  <Td>
                    <Badge colorScheme={getStatusColor(job.status)}>
                      {job.status === 'processing' && <Spinner size="xs" mr={2} />}
                      {job.status}
                    </Badge>
                  </Td>
                  <Td whiteSpace="nowrap">
                    <Tooltip label={`Updated: ${formatDate(job.update_time)}`}>
                      <Text fontSize="sm">{formatDate(job.start_time)}</Text>
                    </Tooltip>
                  </Td>
                  <Td>
                    <Text fontSize="sm" maxW="300px" isTruncated>
                      {job.message || job.stage || 'No message'}
                      {job.error && (
                        <Tooltip label={job.error}>
                          <Icon as={FaExclamationTriangle} color="red.500" ml={2} />
                        </Tooltip>
                      )}
                    </Text>
                  </Td>
                  <Td>
                    <Flex gap={2}>
                      <Tooltip label="View job progress">
                        <Button
                          size="xs"
                          leftIcon={<Icon as={MdVisibility} />}
                          onClick={() => handleSelectJob(job)}
                          colorScheme="blue"
                          variant="outline"
                        >
                          View
                        </Button>
                      </Tooltip>
                      {job.status === 'failed' && (
                        <Tooltip label="Resume failed job">
                          <Button
                            size="xs"
                            leftIcon={<Icon as={FaRedo} />}
                            onClick={() => handleResumeJob(job.job_id)}
                            isLoading={isResuming[job.job_id]}
                            colorScheme="red"
                            variant="outline"
                          >
                            Resume
                          </Button>
                        </Tooltip>
                      )}
                    </Flex>
                  </Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        )}
      </CardBody>

      {/* Job Selection Confirmation Modal */}
      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>
            Job Details
            {selectedJob && (
              <Tag 
                size="sm" 
                colorScheme={getStatusColor(selectedJob.status)} 
                ml={2}
                borderRadius="full"
              >
                {selectedJob.status === 'processing' && (
                  <TagLeftIcon as={Spinner} size="xs" />
                )}
                {selectedJob.status === 'completed' && (
                  <TagLeftIcon as={FaCheck} />
                )}
                {selectedJob.status === 'failed' && (
                  <TagLeftIcon as={FaExclamationTriangle} />
                )}
                <TagLabel>{selectedJob.status}</TagLabel>
              </Tag>
            )}
          </ModalHeader>
          <ModalBody>
            {selectedJob && (
              <Box>
                <Text mb={2}>
                  <strong>Job ID:</strong> {selectedJob.job_id}
                </Text>
                <Text mb={2}>
                  <strong>Started:</strong> {formatDate(selectedJob.start_time)}
                </Text>
                <Text mb={2}>
                  <strong>Last Updated:</strong> {formatDate(selectedJob.update_time)}
                </Text>
                <Text mb={2}>
                  <strong>Progress:</strong> {selectedJob.percent}%
                </Text>
                <Text mb={2}>
                  <strong>Message:</strong> {selectedJob.message}
                </Text>
                {selectedJob.error && (
                  <Text mb={2} color="red.500">
                    <strong>Error:</strong> {selectedJob.error}
                  </Text>
                )}
              </Box>
            )}
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onClose}>
              Cancel
            </Button>
            <Button 
              colorScheme="blue" 
              onClick={handleConfirmJobSelection}
              leftIcon={<Icon as={MdVisibility} />}
            >
              Track This Job
            </Button>
            {selectedJob?.status === 'failed' && (
              <Button 
                colorScheme="red" 
                ml={3}
                onClick={() => {
                  onClose();
                  if (selectedJob) handleResumeJob(selectedJob.job_id);
                }}
                leftIcon={<Icon as={FaRedo} />}
                isLoading={selectedJob ? isResuming[selectedJob.job_id] : false}
              >
                Resume
              </Button>
            )}
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Card>
  );
}