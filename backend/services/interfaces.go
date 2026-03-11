package services

import "aituber/models"

// IScriptGenerator defines the interface for generating scripts
type IScriptGenerator interface {
	GenerateYouTubeScript(topic string) ([]models.VideoSegment, error)
	GenerateTikTokScript(topic string) ([]models.VideoSegment, error)
	HasKeys() bool
}

// IAudioService defines the interface for audio generation and processing
type IAudioService interface {
	GenerateAudioChunks(chunks []string, voice string, speed float64, jobID string, maxConcurrent int) ([]string, error)
	MergeAudioFiles(audioPaths []string, outputPath string) error
}

// IStockVideoService defines the interface for fetching stock clips
type IStockVideoService interface {
	PrepareSegmentVideo(keywords string, audioDuration float64, jobID string, segIndex int, orientation string) (string, error)
}

// IComposerService defines the interface for combining audio and video
type IComposerService interface {
	ComposeVideoWithAudio(videoPath, audioPath, outputPath string) error
}

// IJobManager defines the interface for tracking job progress
type IJobManager interface {
	CreateJob(jobID, platform, contentName string) *models.JobStatus
	GetJob(jobID string) (*models.JobStatus, bool)
	UpdateProgress(jobID string, step string, progress int) error
	MarkFailed(jobID string, err error) error
	MarkCompleted(jobID, videoPath, savedPath string) error
}

// IVideoWorkflow defines the interface for orchestrating video generation
type IVideoWorkflow interface {
	StartGeneration(jobID string, req models.GenerateRequest)
}
