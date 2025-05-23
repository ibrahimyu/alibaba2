package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// YuEConfig holds the configuration for YuE music generation
type YuEConfig struct {
	// Path to the YuE repository
	RepoPath string
	// Path to the stage1 model
	Stage1Model string
	// Path to the stage2 model
	Stage2Model string
	// Path to the checkpoint directory
	CheckpointDir string
	// Path to the output directory
	OutputDir string
	// Number of segments to generate
	RunNSegments int
	// Batch size for stage 2
	Stage2BatchSize int
	// Maximum number of new tokens
	MaxNewTokens int
	// Repetition penalty
	RepetitionPenalty float64
}

// NewDefaultYuEConfig creates a new YuEConfig with default values
func NewDefaultYuEConfig() *YuEConfig {
	return &YuEConfig{
		RepoPath:          os.Getenv("YUE_REPO_PATH"),
		Stage1Model:       "m-a-p/YuE-s1-7B-anneal-en-cot",
		Stage2Model:       "m-a-p/YuE-s2-1B-general",
		CheckpointDir:     os.Getenv("YUE_CHECKPOINT_DIR"),
		OutputDir:         "output",
		RunNSegments:      2,
		Stage2BatchSize:   4,
		MaxNewTokens:      3000,
		RepetitionPenalty: 1.1,
	}
}

// GenerateYuEMusic generates music using the YuE model
func GenerateYuEMusic(genres string, lyrics string, outputDir string) (string, error) {
	// Check if YuE repository path is set
	config := NewDefaultYuEConfig()
	if config.RepoPath == "" {
		return "", fmt.Errorf("YUE_REPO_PATH environment variable is not set")
	}

	// Create the output directory if it doesn't exist
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}

	// Create temporary files for genre and lyrics
	genreFile := filepath.Join(outputDir, "genre.txt")
	lyricsFile := filepath.Join(outputDir, "lyrics.txt")

	// Write genre and lyrics to temporary files
	if err := ioutil.WriteFile(genreFile, []byte(genres), 0644); err != nil {
		return "", fmt.Errorf("failed to write genre file: %w", err)
	}

	// Format the lyrics as required by YuE
	formattedLyrics := formatLyricsForYuE(lyrics)
	if err := ioutil.WriteFile(lyricsFile, []byte(formattedLyrics), 0644); err != nil {
		return "", fmt.Errorf("failed to write lyrics file: %w", err)
	}

	// Set up the YuE command
	inferenceDir := filepath.Join(config.RepoPath, "inference")
	yueOutputDir := filepath.Join(outputDir, "yue_output")

	// Ensure the YuE output directory exists
	if _, err := os.Stat(yueOutputDir); os.IsNotExist(err) {
		os.MkdirAll(yueOutputDir, 0755)
	}

	cmd := exec.Command(
		"python",
		"infer.py",
		"--cuda_idx", "0",
		"--stage1_model", config.Stage1Model,
		"--stage2_model", config.Stage2Model,
		"--genre_txt", genreFile,
		"--lyrics_txt", lyricsFile,
		"--run_n_segments", fmt.Sprintf("%d", config.RunNSegments),
		"--stage2_batch_size", fmt.Sprintf("%d", config.Stage2BatchSize),
		"--output_dir", yueOutputDir,
		"--max_new_tokens", fmt.Sprintf("%d", config.MaxNewTokens),
		"--repetition_penalty", fmt.Sprintf("%f", config.RepetitionPenalty),
	)

	// Set the working directory to the YuE inference directory
	cmd.Dir = inferenceDir

	// Run the command
	log.Printf("Generating music with YuE: %s", cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to generate music with YuE: %w, output: %s", err, string(output))
	}

	log.Printf("YuE output: %s", string(output))

	// Find the generated music file
	entries, err := os.ReadDir(yueOutputDir)
	if err != nil {
		return "", fmt.Errorf("failed to read YuE output directory: %w", err)
	}

	var musicFile string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), "_Mix.wav") {
			musicFile = filepath.Join(yueOutputDir, entry.Name())
			break
		}
	}

	if musicFile == "" {
		return "", fmt.Errorf("no music file found in YuE output directory")
	}

	// Convert WAV to MP3 for better compatibility
	mp3File := strings.TrimSuffix(musicFile, ".wav") + ".mp3"
	ffmpegCmd := exec.Command(
		"ffmpeg",
		"-i", musicFile,
		"-codec:a", "libmp3lame",
		"-qscale:a", "2",
		mp3File,
	)

	ffmpegOutput, err := ffmpegCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to convert WAV to MP3: %w, output: %s", err, string(ffmpegOutput))
	}

	return mp3File, nil
}

// formatLyricsForYuE formats lyrics for YuE
// YuE expects lyrics in a specific format with [verse], [chorus], etc. labels
func formatLyricsForYuE(lyrics string) string {
	if strings.Contains(lyrics, "[verse]") || strings.Contains(lyrics, "[chorus]") {
		// Already formatted for YuE
		return lyrics
	}

	// Simple formatting - split by newlines and wrap as verses
	lines := strings.Split(lyrics, "\n")
	var formattedLyrics strings.Builder

	// First section as verse
	formattedLyrics.WriteString("[verse]\n")
	for i, line := range lines {
		if i > 0 && i%4 == 0 && i < len(lines)-1 {
			// Every 4 lines, start a new section alternating between verse and chorus
			if (i/4)%2 == 0 {
				formattedLyrics.WriteString("\n\n[chorus]\n")
			} else {
				formattedLyrics.WriteString("\n\n[verse]\n")
			}
		}
		formattedLyrics.WriteString(line + "\n")
	}

	return formattedLyrics.String()
}
