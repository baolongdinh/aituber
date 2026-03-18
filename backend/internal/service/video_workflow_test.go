package service

import (
	"aituber/config"
	"aituber/internal/model"
	"aituber/utils"
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func (m *MockJobManager) CreateJob(ctx context.Context, userID, platform, contentName, topic, voice, ttsProvider string) (*model.Job, error) {
	return &model.Job{
		BaseModel:   model.BaseModel{ID: "job1"},
		UserID:      userID,
		Platform:    platform,
		ContentName: contentName,
		Topic:       topic,
		Status:      "processing",
	}, nil
}

func (m *MockJobManager) GetJob(ctx context.Context, jobID string) (*model.Job, error) {
	return &model.Job{
		BaseModel: model.BaseModel{ID: jobID},
		Status:    "processing",
	}, nil
}

func (m *MockJobManager) ListUserJobs(ctx context.Context, userID, platform string, page, limit int) ([]*model.Job, int64, error) {
	return nil, 0, nil
}

func (m *MockJobManager) UpdateProgress(ctx context.Context, jobID, step string, progress int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LastStatus = step
	m.Progress = progress
	return nil
}

func (m *MockJobManager) MarkFailed(ctx context.Context, jobID string, err error) error {
	m.Failed <- err
	return nil
}

func (m *MockJobManager) MarkCompleted(ctx context.Context, jobID, videoPath, savedPath, thumbnailPath string) error {
	m.VideoPath = videoPath
	m.SavedPath = savedPath
	m.Finished <- true
	return nil
}

func (m *MockJobManager) CreateSeries(ctx context.Context, userID, topic, platform, contentName string, numParts int) (*model.Series, error) {
	return &model.Series{
		BaseModel: model.BaseModel{ID: "series1"},
		UserID:    userID,
	}, nil
}

func (m *MockJobManager) GetSeries(ctx context.Context, seriesID string) (*model.Series, error) {
	return nil, nil
}

func (m *MockJobManager) UpdateSeriesStatus(ctx context.Context, seriesID, status string) error {
	return nil
}

func (m *MockJobManager) CreateSeriesPartJob(ctx context.Context, userID, seriesID string, partIndex int, platform, contentName, topic, voice, ttsProvider string) (*model.Job, error) {
	return &model.Job{
		BaseModel: model.BaseModel{ID: "partjob1"},
	}, nil
}

func (m *MockJobManager) GetActiveTask(ctx context.Context, userID, platform string) (*model.Job, *model.Series, error) {
	return nil, nil, nil
}

func (m *MockJobManager) UpdateJobTitle(ctx context.Context, jobID, title string) error {
	return nil
}

type MockGeminiService struct {
	Segments []VideoSegment
	Err      error
}

func (m *MockGeminiService) GenerateYouTubeScript(topic string) (*GeneratedScript, error) {
	return &GeneratedScript{Segments: m.Segments, Title: "AI Title"}, m.Err
}
func (m *MockGeminiService) GenerateTikTokScript(topic string) (*GeneratedScript, error) {
	return &GeneratedScript{Segments: m.Segments, Title: "AI Title"}, m.Err
}
func (m *MockGeminiService) HasKeys() bool { return true }
func (m *MockGeminiService) GenerateSeriesOutline(topic, platform string, numParts int) ([]SeriesPartOutline, error) {
	return nil, nil
}
func (m *MockGeminiService) GenerateSeriesPartScript(topic, platform string, outline []SeriesPartOutline, partIdx int) (*GeneratedScript, error) {
	return &GeneratedScript{Segments: m.Segments, Title: "Part Title"}, m.Err
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
func (m *MockStockVideoService) FetchSourceMaterial(ctx context.Context, keywords string, visualDesc string, t2vModel, t2vProvider string, jobID string, segIndex int, orientation string) (*StockMaterial, error) {
	return &StockMaterial{Type: "image"}, m.Err
}
func (m *MockStockVideoService) PrepareVideoFromMaterial(ctx context.Context, material *StockMaterial, audioDuration float64, jobID string, segIndex int, orientation string) (string, error) {
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

func (m *MockComposerService) ExtractThumbnail(videoPath, outputPath string, timeOffset float64) error {
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
		Segments: []VideoSegment{
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

	req := GenerateRequest{
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
	assert.NoError(t, err)
	assert.NotEmpty(t, srtPath)

	// Verify file content if possible
	content, err := os.ReadFile(srtPath)
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
}
