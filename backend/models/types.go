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
