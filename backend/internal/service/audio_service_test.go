package service

import (
	"aituber/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockVoiceCatalog for testing
type MockVoiceCatalog struct {
	mock.Mock
}

func (m *MockVoiceCatalog) IsSupportedProvider(provider string) bool {
	args := m.Called(provider)
	return args.Bool(0)
}

func (m *MockVoiceCatalog) ValidateProvider(voice, provider string) error {
	args := m.Called(voice, provider)
	return args.Error(0)
}

func (m *MockVoiceCatalog) GetRefAudioURL(voice string) string {
	args := m.Called(voice)
	return args.String(0)
}

func TestMapToElevenLabsVoice(t *testing.T) {
	as := &AudioService{}

	tests := []struct {
		voiceName string
		expected  string // ElevenLabs ID
	}{
		{"minhquang", "ipTvfDXAg1zowfF1rv9w"},                              // Male
		{"giahuy", "ipTvfDXAg1zowfF1rv9w"},                                 // Male
		{"leminh", "Si3s1VCb7dLbeqH57kiC"},                                 // Female (fallback)
		{"random_long_id_already_exists", "random_long_id_already_exists"}, // Pass-through
	}

	for _, tt := range tests {
		result := as.mapToElevenLabsVoice(tt.voiceName)
		if result != tt.expected {
			t.Errorf("mapToElevenLabsVoice(%s) = %s; want %s", tt.voiceName, result, tt.expected)
		}
	}
}

func TestGenerateSingleAudio_HubProvider(t *testing.T) {
	// Setup Hub mock server
	var hubServerURL string
	hubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/generate/tts" {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			var req HubTTSRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "xin chào tôi tên là long", req.Text)
			assert.Equal(t, "f5-tts-vi", req.ModelName)
			assert.Equal(t, "http://localhost:8080/voice/women_north_sound.mp3", req.RefAudioUrl)

			response := HubTTSResponse{
				Status: "success",
				Data: struct {
					JobID  string `json:"job_id"`
					URL    string `json:"url"`
					Status string `json:"status"`
				}{
					JobID:  "test-job-id",
					URL:    hubServerURL + "/audio.wav", // Use captured server URL
					Status: "queued",
				},
				Message: "generation job queued",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else if r.URL.Path == "/audio.wav" {
			// Return mock audio data
			w.Header().Set("Content-Type", "audio/wav")
			w.Write([]byte("FAKE_AUDIO_DATA")) // Mock audio data
		}
	}))
	defer hubServer.Close()
	hubServerURL = hubServer.URL

	// Setup temp directory
	tempDir, err := os.MkdirTemp("", "audio-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create real voice catalog
	voiceCatalog := NewVoiceCatalog()

	// Create audio service using constructor
	apiPool := utils.NewAPIKeyPool([]string{"test-key"})
	audioService := NewAudioService(
		apiPool,
		"test-elevenlabs-key",
		tempDir,
		"128k",
		22050,
		0.1,
		hubServer.URL, // Use full URL with http://
		"test-token",
		"http://localhost:8080", // Base URL for voice files
	)
	audioService.voiceCatalog = voiceCatalog

	// Test Hub TTS generation
	audioPath, err := audioService.GenerateSingleAudio("xin chào tôi tên là long", "banmai", "hub", 1.0, "test-job", 0)

	// Verify the call was successful
	require.NoError(t, err)
	assert.NotEmpty(t, audioPath)
	assert.True(t, strings.HasSuffix(audioPath, "chunk_000.mp3"))
}

func TestGenerateSingleAudio_HubMissingConfig(t *testing.T) {
	// Setup temp directory
	tempDir, err := os.MkdirTemp("", "audio-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create real voice catalog
	voiceCatalog := NewVoiceCatalog()

	// Create audio service using constructor
	apiPool := utils.NewAPIKeyPool([]string{"test-key"})
	audioService := NewAudioService(
		apiPool,
		"test-elevenlabs-key",
		tempDir,
		"128k",
		22050,
		0.1,
		"",                      // No Hub URL
		"",                      // No Hub token
		"http://localhost:8080", // Base URL for voice files
	)
	audioService.voiceCatalog = voiceCatalog

	// Test Hub provider without config
	audioPath, err := audioService.GenerateSingleAudio("test", "banmai", "hub", 1.0, "test-job", 0)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Hub TTS configuration missing")
	assert.Empty(t, audioPath)
}

func TestCallHubTTS_APIRequest(t *testing.T) {
	// Setup Hub mock server
	hubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request details
		assert.Equal(t, "/api/v1/generate/tts", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		var req HubTTSRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "test text", req.Text)
		assert.Equal(t, "f5-tts-vi", req.ModelName)
		assert.Equal(t, "http://example.com/ref.wav", req.RefAudioUrl)

		// Send success response
		response := HubTTSResponse{
			Status: "success",
			Data: struct {
				JobID  string `json:"job_id"`
				URL    string `json:"url"`
				Status string `json:"status"`
			}{
				JobID:  "test-job-123",
				URL:    "http://example.com/generated.wav",
				Status: "queued",
			},
			Message: "generation job queued",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer hubServer.Close()

	// Create audio service using constructor
	apiPool := utils.NewAPIKeyPool([]string{"test-key"})
	audioService := NewAudioService(
		apiPool,
		"test-elevenlabs-key",
		"/tmp",
		"128k",
		22050,
		0.1,
		hubServer.URL, // Use full URL with http://
		"test-token",
		"http://localhost:8080", // Base URL for voice files
	)

	// Test the Hub API call
	audioURL, err := audioService.callHubTTS("test text", "http://example.com/ref.wav")

	require.NoError(t, err)
	assert.Equal(t, "http://example.com/generated.wav", audioURL)
}

func TestHubTTSIntegration_FullWorkflow(t *testing.T) {
	// This test simulates the full workflow from frontend to Hub TTS

	// 1. Setup Hub mock server with realistic response
	hubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req HubTTSRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Simulate realistic Hub TTS response
		response := HubTTSResponse{
			Status: "success",
			Data: struct {
				JobID  string `json:"job_id"`
				URL    string `json:"url"`
				Status string `json:"status"`
			}{
				JobID:  "04c88fe5-b923-431d-9ecd-bd26215420d6",
				URL:    "http://example.com/generated.wav", // Fixed URL
				Status: "queued",
			},
			Message: "[f5-tts-vi, tts] generation job queued: 04c88fe5-b923-431d-9ecd-bd26215420d6",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer hubServer.Close()

	// 2. Setup temp directory
	tempDir, err := os.MkdirTemp("", "audio-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 3. Create real voice catalog
	voiceCatalog := NewVoiceCatalog()

	// 4. Create audio service with Hub configuration
	apiPool := utils.NewAPIKeyPool([]string{"test-key"})
	audioService := NewAudioService(
		apiPool,
		"test-elevenlabs-key",
		tempDir,
		"128k",
		22050,
		0.1,
		hubServer.URL, // Use full URL with http://
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJsdWFuYWlveiJ9.E6aKblsqO9SovQAv4wDgD259iFYTSvp3xHjJoZHuUEo",
		"http://localhost:8080", // Base URL for voice files
	)
	audioService.voiceCatalog = voiceCatalog

	// 5. Test the complete workflow
	audioPath, err := audioService.GenerateSingleAudio("xin chào tôi tên là long", "banmai", "hub", 1.0, "test-job", 0)

	// 6. Verify results
	require.NoError(t, err)
	assert.NotEmpty(t, audioPath)
	assert.True(t, strings.HasSuffix(audioPath, "chunk_000.mp3"))

	// 7. Verify file was created
	_, err = os.Stat(audioPath)
	assert.NoError(t, err)
}
