package services

import (
	"aituber/config"
	"aituber/models"
	"testing"
)

// --- MOCK DEFINITIONS ---

type MockJobManager struct{}

func (m *MockJobManager) CreateJob(jobID, platform, contentName string) *models.JobStatus {
	return &models.JobStatus{JobID: jobID, Platform: platform}
}
func (m *MockJobManager) GetJob(jobID string) (*models.JobStatus, bool) {
	return &models.JobStatus{JobID: jobID}, true
}
func (m *MockJobManager) UpdateProgress(jobID string, step string, progress int) error { return nil }
func (m *MockJobManager) MarkFailed(jobID string, err error) error                     { return nil }
func (m *MockJobManager) MarkCompleted(jobID, videoPath, savedPath string) error       { return nil }

type MockGeminiService struct {
	Segments []models.VideoSegment
	Err      error
}

func (m *MockGeminiService) GenerateYouTubeScript(topic string) ([]models.VideoSegment, error) {
	return m.Segments, m.Err
}
func (m *MockGeminiService) GenerateTikTokScript(topic string) ([]models.VideoSegment, error) {
	return m.Segments, m.Err
}
func (m *MockGeminiService) HasKeys() bool { return true }

type MockAudioService struct {
	AudioPaths []string
	Err        error
}

func (m *MockAudioService) GenerateAudioChunks(chunks []string, voice string, speed float64, jobID string, maxConcurrent int) ([]string, error) {
	return m.AudioPaths, m.Err
}
func (m *MockAudioService) MergeAudioFiles(audioPaths []string, outputPath string) error {
	return m.Err
}

type MockStockVideoService struct {
	VideoPath string
	Err       error
}

func (m *MockStockVideoService) PrepareSegmentVideo(keywords string, audioDuration float64, jobID string, segIndex int, orientation string) (string, error) {
	return m.VideoPath, m.Err
}

type MockComposerService struct {
	Err error
}

func (m *MockComposerService) ComposeVideoWithAudio(videoPath, audioPath, outputPath string) error {
	return m.Err
}

// --- TESTS ---

func TestVideoWorkflowService_StartGeneration(t *testing.T) {
	cfg := &config.Config{
		TempDir:              "/tmp",
		OutputDir:            "/output",
		MaxTextLength:        1000,
		VideoSegmentDuration: 5.0,
	}

	jm := &MockJobManager{}
	tp := NewTextProcessor(1000, 5.0)

	gemini := &MockGeminiService{
		Segments: []models.VideoSegment{
			{Text: "Test segment 1", VisualPrompt: "nature"},
		},
	}

	audio := &MockAudioService{
		AudioPaths: []string{"/tmp/audio1.mp3"},
	}

	stock := &MockStockVideoService{
		VideoPath: "/tmp/video1.mp4",
	}

	composer := &MockComposerService{}

	// videoService is not using interface yet, but it's okay for now as most logic is in workflow
	// If we need to mock it, we'll need another interface.
	workflow := NewVideoWorkflowService(cfg, jm, tp, audio, nil, stock, composer, gemini)

	req := models.GenerateRequest{
		Topic:    "Test Topic",
		Platform: "tiktok",
	}

	// We can't easily wait for 'go' routine in StartGeneration without adding synchronization or mocking the orchestrator differently.
	// For unit test, we might want to test the sub-functions or refactor StartGeneration to be more testable (return error or take a channel).

	// Let's test sub-pipelines directly for now to ensure logic is sound.
	t.Run("GenerateScript", func(t *testing.T) {
		segments, err := workflow.generateScript("job1", req)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(segments) != 1 {
			t.Errorf("Expected 1 segment, got %d", len(segments))
		}
	})

	t.Run("GenerateAudio", func(t *testing.T) {
		segments := []models.VideoSegment{{Text: "Hello"}}
		paths, texts, err := workflow.generateAudio("job1", req, segments)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(paths) != 1 || len(texts) != 1 {
			t.Errorf("Expected 1 path/text, got %d/%d", len(paths), len(texts))
		}
	})
}
