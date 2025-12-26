package models

import "time"

// GenerateRequest represents the input from frontend
type GenerateRequest struct {
	Script        string  `json:"script" binding:"required"`
	Voice         string  `json:"voice" binding:"required"`
	SpeakingSpeed float64 `json:"speaking_speed"`
	VideoStyle    string  `json:"video_style"`
	VideoSource   string  `json:"video_source"`   // "ai" or "stock"
	StockKeywords string  `json:"stock_keywords"` // Keywords for Pexels search
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
	Error       *string `json:"error,omitempty"`
}

// VideoSegment represents a text segment with duration
type VideoSegment struct {
	Text              string
	EstimatedDuration float64
	VisualPrompt      string
}

// JobStatus tracks processing status in memory
type JobStatus struct {
	JobID       string
	Status      string
	Progress    int
	CurrentStep string
	VideoPath   string
	Error       error
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
