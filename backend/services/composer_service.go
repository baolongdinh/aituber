package services

import (
	"aituber/utils"
	"fmt"
)

// ComposerService combines audio and video into final output
type ComposerService struct {
	videoBitrate string
}

// NewComposerService creates a new composer service
func NewComposerService(videoBitrate string) *ComposerService {
	return &ComposerService{
		videoBitrate: videoBitrate,
	}
}

// ComposeVideoWithAudio combines video and audio tracks
func (cs *ComposerService) ComposeVideoWithAudio(videoPath, audioPath, outputPath string) error {
	if videoPath == "" || audioPath == "" {
		return fmt.Errorf("video and audio paths are required")
	}

	// Use FFmpeg utility to combine
	err := utils.CombineAudioVideo(videoPath, audioPath, outputPath, cs.videoBitrate)
	if err != nil {
		return fmt.Errorf("failed to compose video: %w", err)
	}

	return nil
}
