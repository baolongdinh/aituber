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
	Port     string
	TempDir  string
	CacheDir string

	// Output directory for saved videos
	OutputDir string

	// Subtitles
	EnableSubtitles bool

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// JWT Authentication
	JWTSecret            string
	JWTAccessTokenExpiry int // hours

	// API Keys Pool
	TTSAPIKeys       []string
	ElevenLabsAPIKey string
	VideoAPIKeys     []string
	GeminiAPIKeys    []string
	LocalHubURL      string
	RemoteHubURL     string
	RemoteHubToken   string

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

	PexelsAPIKey      string
	HuggingFaceTokens []string

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
		Port:      getEnv("PORT", "8080"),
		TempDir:   getEnv("TEMP_DIR", "./temp"),
		OutputDir: getEnv("OUTPUT_DIR", "../ai-videos"),
		CacheDir:  getEnv("CACHE_DIR", "./cache"),

		EnableSubtitles: getEnvAsBool("ENABLE_SUBTITLES", false),

		// Database
		DBHost:     getEnv("DATABASE_HOST", "localhost"),
		DBPort:     getEnv("DATABASE_PORT", "5432"),
		DBUser:     getEnv("DATABASE_USER", "postgres"),
		DBPassword: getEnv("DATABASE_PASSWORD", ""),
		DBName:     getEnv("DATABASE_NAME", "aituber"),

		// JWT
		JWTSecret:            getEnv("JWT_SECRET", "change-this-secret-in-production"),
		JWTAccessTokenExpiry: getEnvAsInt("JWT_ACCESS_TOKEN_EXPIRY_HOURS", 24),

		// Parse API keys
		TTSAPIKeys:       parseAPIKeys(getEnv("TTS_API_KEYS", "")),
		ElevenLabsAPIKey: getEnv("ELEVENLABS_API_KEY", ""),
		VideoAPIKeys:     parseAPIKeys(getEnv("VIDEO_API_KEYS", "")),
		GeminiAPIKeys:    parseAPIKeys(getEnv("GEMINI_API_KEYS", "")),
		LocalHubURL:      getEnv("LOCAL_HUB_URL", "http://localhost:5000"),
		RemoteHubURL:     getEnv("REMOTE_HUB_URL", "http://10.0.0.224:8081"),
		RemoteHubToken:   getEnv("REMOTE_HUB_TOKEN", ""),

		// Processing settings
		MaxTextLength:        getEnvAsInt("MAX_TEXT_LENGTH", 50000),
		AudioChunkSize:       getEnvAsInt("AUDIO_CHUNK_SIZE", 8000),
		VideoSegmentDuration: getEnvAsFloat("VIDEO_SEGMENT_DURATION", 10.0),

		// Quality settings
		AudioSampleRate: getEnvAsInt("AUDIO_SAMPLE_RATE", 44100),
		AudioBitrate:    getEnv("AUDIO_BITRATE", "320k"),
		VideoBitrate:    getEnv("VIDEO_BITRATE", "8M"),
		VideoResolution: getEnv("VIDEO_RESOLUTION", "1920x1080"),
		VideoFPS:        getEnvAsInt("VIDEO_FPS", 30),

		// Transition settings
		AudioCrossfadeDuration:  getEnvAsFloat("AUDIO_CROSSFADE_DURATION", 0.0),
		VideoTransitionType:     getEnv("VIDEO_TRANSITION_TYPE", "fade"),
		VideoTransitionDuration: getEnvAsFloat("VIDEO_TRANSITION_DURATION", 0.5),

		PexelsAPIKey:      getEnv("PEXELS_API_KEY", ""),
		HuggingFaceTokens: parseAPIKeys(getEnv("HF_TOKEN", "")),

		// Rate limiting
		MaxConcurrentTTSRequests:   getEnvAsInt("MAX_CONCURRENT_TTS_REQUESTS", 1),
		MaxConcurrentVideoRequests: getEnvAsInt("MAX_CONCURRENT_VIDEO_REQUESTS", 5),
		RetryDelaySeconds:          getEnvAsInt("RETRY_DELAY_SECONDS", 60),
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetDatabaseDSN returns the Data Source Name for PostgreSQL
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort)
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	// If using FPT TTS, require API keys. If using Hub only, FPT keys not required.
	if len(c.TTSAPIKeys) == 0 && c.RemoteHubURL == "" {
		return errors.New("either TTS_API_KEYS (for FPT) or REMOTE_HUB_URL (for Hub) is required")
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

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
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
	return fmt.Sprintf("Config{Port: %s, TTS Keys: %d, Gemini Keys: %d, ChunkSize: %d, OutputDir: %s}",
		c.Port, len(c.TTSAPIKeys), len(c.GeminiAPIKeys), c.AudioChunkSize, c.OutputDir)
}
