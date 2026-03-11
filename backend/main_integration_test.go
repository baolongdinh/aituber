package main

import (
	"aituber/config"
	"aituber/handlers"
	"aituber/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// --- MOCK DEFINITIONS ---

type MockJobManagerIntegration struct {
	LastJobID string
}

func (m *MockJobManagerIntegration) CreateJob(jobID, platform, contentName string) *models.JobStatus {
	m.LastJobID = jobID
	return &models.JobStatus{JobID: jobID}
}
func (m *MockJobManagerIntegration) GetJob(jobID string) (*models.JobStatus, bool) {
	if jobID == "test-123" {
		return &models.JobStatus{JobID: jobID, Status: "completed", Progress: 100}, true
	}
	return nil, false
}
func (m *MockJobManagerIntegration) UpdateProgress(jobID string, step string, progress int) error {
	return nil
}
func (m *MockJobManagerIntegration) MarkFailed(jobID string, err error) error { return nil }
func (m *MockJobManagerIntegration) MarkCompleted(jobID, videoPath, savedPath string) error {
	return nil
}

type MockWorkflowIntegration struct {
	Started bool
}

func (m *MockWorkflowIntegration) StartGeneration(jobID string, req models.GenerateRequest) {
	m.Started = true
}

type MockGeminiIntegration struct{}

func (m *MockGeminiIntegration) GenerateYouTubeScript(topic string) ([]models.VideoSegment, error) {
	return []models.VideoSegment{{Text: "Test segment"}}, nil
}
func (m *MockGeminiIntegration) GenerateTikTokScript(topic string) ([]models.VideoSegment, error) {
	return []models.VideoSegment{{Text: "Test segment"}}, nil
}
func (m *MockGeminiIntegration) HasKeys() bool { return true }
func (m *MockGeminiIntegration) GenerateSeriesOutline(topic, platform string, numParts int) ([]models.SeriesPartOutline, error) {
	return nil, nil
}
func (m *MockGeminiIntegration) GenerateSeriesPartScript(topic, platform string, outline []models.SeriesPartOutline, partIdx int) ([]models.VideoSegment, error) {
	return nil, nil
}

func TestAPI_Integration_FullFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	cfg := &config.Config{
		TempDir:   "/tmp",
		OutputDir: "/output",
	}

	jm := &MockJobManagerIntegration{}
	gemini := &MockGeminiIntegration{}
	workflow := &MockWorkflowIntegration{}

	h := handlers.NewVideoHandler(cfg, jm, workflow, gemini)

	api := router.Group("/api")
	{
		api.POST("/generate", h.Generate)       // In video_handler.go it is named 'Generate'
		api.GET("/status/:job_id", h.GetStatus) // In video_handler.go it is named 'GetStatus'
	}

	t.Run("POST /api/generate", func(t *testing.T) {
		reqBody := models.GenerateRequest{
			Topic:    "Integration Test",
			Platform: "youtube",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/api/generate", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp models.GenerateResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.JobID == "" {
			t.Error("Expected JobID to be returned")
		}
		if !workflow.Started {
			t.Error("Expected workflow to be started")
		}
	})

	t.Run("GET /api/status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/status/test-123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp models.StatusResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Status != "completed" {
			t.Errorf("Expected status completed, got %s", resp.Status)
		}
	})
}
