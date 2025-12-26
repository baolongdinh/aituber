package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port    string
	TempDir string

	// API Keys Pool
	TTSAPIKeys   []string
	VideoAPIKeys []string

	// Processing Settings
	MaxTextLength        int
	AudioChunkSize       int
	VideoSegmentDuration float64

	// Quality Settings
	AudioSampleRate int
	AudioBitrate    string
	VideoBitrate    string
	VideoResolution string
	VideoFPS        int

	// Transition Settings
	AudioCrossfadeDuration  float64
	VideoTransitionType     string
	VideoTransitionDuration float64

	PexelsAPIKey string

	// Rate Limiting
	MaxConcurrentTTSRequests   int
	MaxConcurrentVideoRequests int
	RetryDelaySeconds          int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	cfg := &Config{
		Port:    getEnv("PORT", "8080"),
		TempDir: getEnv("TEMP_DIR", "./backend/temp"),

		// Parse API keys
		TTSAPIKeys:   parseAPIKeys(getEnv("TTS_API_KEYS", "")),
		VideoAPIKeys: parseAPIKeys(getEnv("VIDEO_API_KEYS", "")),

		// Processing settings
		MaxTextLength:        getEnvAsInt("MAX_TEXT_LENGTH", 50000),
		AudioChunkSize:       getEnvAsInt("AUDIO_CHUNK_SIZE", 4500),
		VideoSegmentDuration: getEnvAsFloat("VIDEO_SEGMENT_DURATION", 5.5),

		// Quality settings
		AudioSampleRate: getEnvAsInt("AUDIO_SAMPLE_RATE", 44100),
		AudioBitrate:    getEnv("AUDIO_BITRATE", "192k"),
		VideoBitrate:    getEnv("VIDEO_BITRATE", "5M"),
		VideoResolution: getEnv("VIDEO_RESOLUTION", "1920x1080"),
		VideoFPS:        getEnvAsInt("VIDEO_FPS", 30),

		// Transition settings
		AudioCrossfadeDuration:  getEnvAsFloat("AUDIO_CROSSFADE_DURATION", 0.3),
		VideoTransitionType:     getEnv("VIDEO_TRANSITION_TYPE", "fade"),
		VideoTransitionDuration: getEnvAsFloat("VIDEO_TRANSITION_DURATION", 0.5),

		PexelsAPIKey: getEnv("PEXELS_API_KEY", ""),

		// Rate limiting
		MaxConcurrentTTSRequests:   getEnvAsInt("MAX_CONCURRENT_TTS_REQUESTS", 3),
		MaxConcurrentVideoRequests: getEnvAsInt("MAX_CONCURRENT_VIDEO_REQUESTS", 2),
		RetryDelaySeconds:          getEnvAsInt("RETRY_DELAY_SECONDS", 60),
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	if len(c.TTSAPIKeys) == 0 {
		return errors.New("TTS_API_KEYS is required")
	}
	if len(c.VideoAPIKeys) == 0 {
		return errors.New("VIDEO_API_KEYS is required")
	}
	if c.AudioChunkSize <= 0 {
		return errors.New("AUDIO_CHUNK_SIZE must be positive")
	}
	if c.VideoSegmentDuration <= 0 {
		return errors.New("VIDEO_SEGMENT_DURATION must be positive")
	}
	return nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func parseAPIKeys(keysStr string) []string {
	if keysStr == "" {
		return []string{}
	}
	keys := strings.Split(keysStr, ",")
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		trimmed := strings.TrimSpace(key)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{Port: %s, TTS Keys: %d, Video Keys: %d, ChunkSize: %d}",
		c.Port, len(c.TTSAPIKeys), len(c.VideoAPIKeys), c.AudioChunkSize)
}
