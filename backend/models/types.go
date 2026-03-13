package models

import "time"

// GenerateRequest represents the input from frontend
type GenerateRequest struct {
	// Platform: "youtube" or "tiktok"
	Platform string `json:"platform" binding:"required"`
	// Topic: what the video is about (AI will generate the script)
	Topic string `json:"topic" binding:"required"`
	// ContentName: optional folder name for output (auto-generated from topic if empty)
	ContentName string `json:"content_name"`

	// Audio settings
	Voice         string  `json:"voice" binding:"required"`
	SpeakingSpeed float64 `json:"speaking_speed"`

	// Legacy / optional: pre-written script (bypasses Gemini gen if provided)
	Script        string `json:"script"`
	VideoStyle    string `json:"video_style"`
	VideoSource   string `json:"video_source"`
	StockKeywords string `json:"stock_keywords"`
	TTSProvider   string `json:"tts_provider"` // "fpt" or "elevenlabs"
	T2VModel      string `json:"t2v_model"`    // e.g. "genmo/mochi-1-preview"
	T2VProvider   string `json:"t2v_provider"` // e.g. "fal-ai"

	// If Segments is provided, it bypasses both Script text and AI generation
	Segments []VideoSegment `json:"segments"`
}

// GenerateResponse returns the job ID
type GenerateResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

// StatusResponse returns current progress
type StatusResponse struct {
	Status      string  `json:"status"` // "processing", "completed", "failed"
	Progress    int     `json:"progress"`
	CurrentStep string  `json:"current_step"`
	VideoURL    *string `json:"video_url,omitempty"`
	SavedPath   *string `json:"saved_path,omitempty"`
	Error       *string `json:"error,omitempty"`
}

// VideoSegment represents a text segment with duration
type VideoSegment struct {
	Text              string  `json:"text"`
	EstimatedDuration float64 `json:"estimated_duration,omitempty"`
	VisualPrompt      string  `json:"pexels_search_query"`
	VisualDescription string  `json:"visual_description"`
}

// JobStatus tracks processing status in memory
type JobStatus struct {
	JobID       string
	Platform    string
	ContentName string
	Status      string
	Progress    int
	CurrentStep string
	VideoPath   string
	SavedPath   string
	Error       error
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// StockMaterial represents intermediate fetched material (image or video metadata)
type StockMaterial struct {
	Type       string      // "image", "video", "pexels"
	ImageBytes []byte      // Raw bytes for AI images
	VideoPath  string      // Path to downloaded T2V video
	PexelsInfo []VideoInfo // List of pexels clip info
}

type VideoInfo struct {
	Link     string
	Duration int
}

// ---------- Series Video Generation ----------

// SeriesGenerateRequest – POST /api/generate-series
type SeriesGenerateRequest struct {
	Platform      string  `json:"platform" binding:"required"` // "youtube" | "tiktok"
	Topic         string  `json:"topic" binding:"required"`
	NumParts      int     `json:"num_parts" binding:"required"` // 2 – 20
	Voice         string  `json:"voice" binding:"required"`
	SpeakingSpeed float64 `json:"speaking_speed"`
	ContentName   string  `json:"content_name"` // optional slug
	TTSProvider   string  `json:"tts_provider"` // "fpt" or "elevenlabs"
	T2VModel      string  `json:"t2v_model"`    // e.g. "genmo/mochi-1-preview"
	T2VProvider   string  `json:"t2v_provider"` // e.g. "fal-ai"
}

// SeriesGenerateResponse – returned immediately after POST
type SeriesGenerateResponse struct {
	SeriesID string `json:"series_id"`
	Status   string `json:"status"`
	NumParts int    `json:"num_parts"`
}

// SeriesPartStatus – status of one part inside a series
type SeriesPartStatus struct {
	PartIndex   int     `json:"part_index"` // 0-based
	Title       string  `json:"title"`
	Status      string  `json:"status"` // "queued" | "processing" | "completed" | "failed"
	Progress    int     `json:"progress"`
	CurrentStep string  `json:"current_step,omitempty"`
	VideoURL    *string `json:"video_url,omitempty"`
	SavedPath   *string `json:"saved_path,omitempty"`
	Error       *string `json:"error,omitempty"`
}

// SeriesJobStatus – in-memory tracker for the whole series
type SeriesJobStatus struct {
	SeriesID      string
	Topic         string
	NumParts      int
	Platform      string
	ContentName   string
	Voice         string
	SpeakingSpeed float64
	TTSProvider   string
	T2VModel      string
	T2VProvider   string
	Status        string // "processing" | "completed" | "partial_failed" | "failed"
	Parts         []*SeriesPartStatus
	Scripts       [][]VideoSegment // Persisted scripts for each part index
	ChildJobIDs   []string         // jobID per part (reuses existing JobStatus)
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SeriesPartOutline – one element from the Gemini series outline
type SeriesPartOutline struct {
	PartNumber int      `json:"part_number"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	KeyPoints  []string `json:"key_points"`
}
