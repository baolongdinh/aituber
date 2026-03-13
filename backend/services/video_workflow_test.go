package services

import (
	"aituber/config"
	"aituber/models"
	"aituber/utils"
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// --- MOCK DEFINITIONS ---

type MockJobManager struct {
	Finished   chan bool
	Failed     chan error
	VideoPath  string
	SavedPath  string
	LastStatus string
	Progress   int
	mu         sync.Mutex
}

func (m *MockJobManager) CreateJob(jobID, platform, contentName string) *models.JobStatus {
	return &models.JobStatus{JobID: jobID, Platform: platform}
}
func (m *MockJobManager) GetJob(jobID string) (*models.JobStatus, bool) {
	return &models.JobStatus{JobID: jobID}, true
}
func (m *MockJobManager) UpdateProgress(jobID string, step string, progress int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LastStatus = step
	m.Progress = progress
	return nil
}
func (m *MockJobManager) MarkFailed(jobID string, err error) error {
	m.Failed <- err
	return nil
}
func (m *MockJobManager) MarkCompleted(jobID, videoPath, savedPath string) error {
	m.VideoPath = videoPath
	m.SavedPath = savedPath
	m.Finished <- true
	return nil
}

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
func (m *MockGeminiService) GenerateSeriesOutline(topic, platform string, numParts int) ([]models.SeriesPartOutline, error) {
	return nil, nil
}
func (m *MockGeminiService) GenerateSeriesPartScript(topic, platform string, outline []models.SeriesPartOutline, partIdx int) ([]models.VideoSegment, error) {
	return nil, nil
}

type MockAudioService struct {
	AudioPaths []string
	Err        error
}

func (m *MockAudioService) GenerateAudioChunks(chunks []string, voice string, speed float64, jobID string, maxConcurrent int) ([]string, error) {
	return m.AudioPaths, m.Err
}
func (m *MockAudioService) GenerateSingleAudio(text, voice string, speed float64, jobID string, index int) (string, error) {
	if index < len(m.AudioPaths) {
		return m.AudioPaths[index], m.Err
	}
	return "fake_audio.mp3", m.Err // Fallback for testing
}
func (m *MockAudioService) MergeAudioFiles(audioPaths []string, outputPath string) error {
	return m.Err
}

type MockStockVideoService struct {
	VideoPath string
	Err       error
}

func (m *MockStockVideoService) PrepareSegmentVideo(ctx context.Context, keywords string, visualDesc string, t2vModel, t2vProvider string, audioDuration float64, jobID string, segIndex int, orientation string) (string, error) {
	return m.VideoPath, m.Err
}
func (m *MockStockVideoService) FetchSourceMaterial(ctx context.Context, keywords string, visualDesc string, t2vModel, t2vProvider string, jobID string, segIndex int, orientation string) (*models.StockMaterial, error) {
	return &models.StockMaterial{Type: "image"}, m.Err
}
func (m *MockStockVideoService) PrepareVideoFromMaterial(ctx context.Context, material *models.StockMaterial, audioDuration float64, jobID string, segIndex int, orientation string) (string, error) {
	return m.VideoPath, m.Err
}

type MockComposerService struct {
	Err error
}

func (m *MockComposerService) ComposeVideoWithAudio(videoPath, audioPath, outputPath string) error {
	return m.Err
}

func (m *MockComposerService) ConcatVideos(videoPaths []string, outputPath string) error {
	return m.Err
}

// --- TESTS ---

func TestVideoWorkflowService_StartGeneration_Success(t *testing.T) {
	// Setup mocks
	originalGetDuration := utils.GetDurationFunc
	utils.GetDurationFunc = func(path string) (float64, error) {
		return 10.0, nil
	}
	defer func() { utils.GetDurationFunc = originalGetDuration }()

	tempDir, _ := os.MkdirTemp("", "workflow_test")
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		TempDir:                  tempDir,
		OutputDir:                tempDir,
		MaxTextLength:            1000,
		VideoSegmentDuration:     5.0,
		MaxConcurrentTTSRequests: 2,
	}

	jm := &MockJobManager{
		Finished: make(chan bool, 1),
		Failed:   make(chan error, 1),
	}
	tp := NewTextProcessor(1000, 5.0)

	gemini := &MockGeminiService{
		Segments: []models.VideoSegment{
			{Text: "Test segment 1", VisualPrompt: "nature"},
			{Text: "Test segment 2", VisualPrompt: "tech"},
		},
	}

	audio := &MockAudioService{
		AudioPaths: []string{
			filepath.Join(tempDir, "audio0.mp3"),
			filepath.Join(tempDir, "audio1.mp3"),
		},
	}
	// Create dummy audio files
	for _, p := range audio.AudioPaths {
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte("fake audio"), 0644)
	}

	stock := &MockStockVideoService{
		VideoPath: filepath.Join(tempDir, "seg.mp4"),
	}
	os.WriteFile(stock.VideoPath, []byte("fake video"), 0644)

	composer := &MockComposerService{}

	workflow := NewVideoWorkflowService(cfg, jm, tp, audio, nil, stock, composer, gemini)

	req := models.GenerateRequest{
		Topic:    "Test Topic",
		Platform: "tiktok",
	}

	// We run it because StartGeneration itself is the orchestrator.
	// In the real app, it's called as `go workflow.StartGeneration(...)`.
	// We'll call it here and wait on the Finished channel.
	go workflow.StartGeneration("job1", req)

	select {
	case <-jm.Finished:
		// Success
		t.Log("Workflow finished successfully")
	case err := <-jm.Failed:
		t.Errorf("Workflow failed unexpectedly: %v", err)
	case <-time.After(10 * time.Second):
		t.Error("Workflow timed out")
	}
}

func TestVideoWorkflowService_GenerateSRT(t *testing.T) {
	// Setup mocks
	originalGetDuration := utils.GetDurationFunc
	utils.GetDurationFunc = func(path string) (float64, error) {
		return 10.0, nil
	}
	defer func() { utils.GetDurationFunc = originalGetDuration }()

	tempDir, _ := os.MkdirTemp("", "srt_test")
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		TempDir: tempDir,
	}
	workflow := NewVideoWorkflowService(cfg, nil, nil, nil, nil, nil, nil, nil)

	audioPaths := []string{
		filepath.Join(tempDir, "a1.mp3"),
		filepath.Join(tempDir, "a2.mp3"),
	}
	texts := []string{"Subtitle 1", "Subtitle 2"}

	for _, p := range audioPaths {
		os.WriteFile(p, []byte("fake"), 0644)
	}

	srtPath, err := workflow.GenerateSRT("job1", audioPaths, texts, tempDir, "tiktok")
	if err != nil {
		t.Fatalf("GenerateSRT failed: %v", err)
	}

	if srtPath == "" {
		t.Error("Expected SRT path, got empty string")
	}

	// Verify file content if possible
	content, err := os.ReadFile(srtPath)
	if err != nil {
		t.Errorf("Failed to read SRT file: %v", err)
	}
	if len(content) == 0 {
		t.Error("SRT file is empty")
	}
}
