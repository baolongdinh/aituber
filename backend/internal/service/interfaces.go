package service

import (
	"aituber/internal/model"
	"context"
)

// --- Types ---

type VideoSegment struct {
	Text              string  `json:"text"`
	EstimatedDuration float64 `json:"estimated_duration,omitempty"`
	VisualPrompt      string  `json:"pexels_search_query"`
	VisualDescription string  `json:"visual_description"`
}

type GeneratedScript struct {
	Title    string         `json:"title"`
	Segments []VideoSegment `json:"segments"`
}

type GenerateRequest struct {
	Platform      string         `json:"platform"`
	Topic         string         `json:"topic"`
	ContentName   string         `json:"content_name"`
	Voice         string         `json:"voice"`
	SpeakingSpeed float64        `json:"speaking_speed"`
	TTSProvider   string         `json:"tts_provider"`
	T2VModel      string         `json:"t2v_model"`
	T2VProvider   string         `json:"t2v_provider"`
	StockKeywords string         `json:"stock_keywords"`
	Script        string         `json:"script"`
	Segments      []VideoSegment `json:"segments,omitempty"`
}

type SeriesGenerateRequest struct {
	Platform      string  `json:"platform"`
	Topic         string  `json:"topic"`
	ContentName   string  `json:"content_name"`
	NumParts      int     `json:"num_parts"`
	Voice         string  `json:"voice"`
	SpeakingSpeed float64 `json:"speaking_speed"`
	TTSProvider   string  `json:"tts_provider"`
	T2VModel      string  `json:"t2v_model"`
	T2VProvider   string  `json:"t2v_provider"`
}

type SeriesPartOutline struct {
	PartNumber int      `json:"part_number"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	KeyPoints  []string `json:"key_points"`
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

// --- Interfaces ---

// IVideoWorkflow defines the interface for orchestrating video generation
type IVideoWorkflow interface {
	StartGeneration(jobID string, req GenerateRequest)
	CancelJob(jobID string) bool
}

// IScriptGenerator defines the interface for generating scripts using AI
type IScriptGenerator interface {
	GenerateYouTubeScript(topic string) (*GeneratedScript, error)
	GenerateTikTokScript(topic string) (*GeneratedScript, error)
	GenerateSeriesOutline(topic, platform string, numParts int) ([]SeriesPartOutline, error)
	GenerateSeriesPartScript(topic, platform string, outline []SeriesPartOutline, partIdx int) (*GeneratedScript, error)
	HasKeys() bool
}

// IAudioService defines the interface for audio generation and processing
type IAudioService interface {
	GenerateAudioChunks(chunks []string, voice string, speed float64, jobID string, maxConcurrent int) ([]string, error)
	GenerateSingleAudio(text, voice, provider string, speed float64, jobID string, index int) (string, error)
	MergeAudioFiles(audioPaths []string, outputPath string) error
}

// IStockVideoService defines the interface for fetching stock clips
type IStockVideoService interface {
	PrepareSegmentVideo(ctx context.Context, keywords string, visualDesc string, t2vModel, t2vProvider string, audioDuration float64, jobID string, segIndex int, orientation string) (string, error)
	FetchSourceMaterial(ctx context.Context, keywords string, visualDesc string, t2vModel, t2vProvider string, jobID string, segIndex int, orientation string) (*StockMaterial, error)
	PrepareVideoFromMaterial(ctx context.Context, material *StockMaterial, audioDuration float64, jobID string, segIndex int, orientation string) (string, error)
}

// IComposerService defines the interface for combining audio and video
type IComposerService interface {
	ComposeVideoWithAudio(videoPath, audioPath, outputPath string) error
	ConcatVideos(videoPaths []string, outputPath string) error
	ExtractThumbnail(videoPath, outputPath string, timeOffset float64) error
}

// IVideoProcessor handles low level video manipulation (renamed from legacy VideoService)
type IVideoProcessor interface {
	GenerateVideoPrompts(segments []VideoSegment, style string) ([]string, error)
	GenerateVideos(prompts []string, durations []float64, jobID string, maxConcurrent int) ([]string, error)
	MergeVideos(videoPaths []string, outputPath string) error
}

// JobService handles video generation job lifecycle and persistence
type JobService interface {
	CreateJob(ctx context.Context, userID, platform, contentName, topic, voice, ttsProvider string) (*model.Job, error)
	GetJob(ctx context.Context, jobID string) (*model.Job, error)
	ListUserJobs(ctx context.Context, userID, platform string, page, limit int) ([]*model.Job, int64, error)
	UpdateProgress(ctx context.Context, jobID, step string, progress int) error
	MarkFailed(ctx context.Context, jobID string, err error) error
	MarkCompleted(ctx context.Context, jobID, videoPath, savedPath, thumbnailPath string) error
	CreateSeries(ctx context.Context, userID, topic, platform, contentName string, numParts int) (*model.Series, error)
	GetSeries(ctx context.Context, seriesID string) (*model.Series, error)
	UpdateSeriesStatus(ctx context.Context, seriesID, status string) error
	CreateSeriesPartJob(ctx context.Context, userID, seriesID string, partIndex int, platform, contentName, topic, voice, ttsProvider string) (*model.Job, error)
	GetActiveTask(ctx context.Context, userID, platform string) (*model.Job, *model.Series, error)
	UpdateJobTitle(ctx context.Context, jobID, title string) error
	SaveCheckpoint(ctx context.Context, jobID string, checkpoint *model.JobCheckpoint) error
	GetCheckpoint(ctx context.Context, jobID string) (*model.JobCheckpoint, error)
}

// VideoService handles video gallery and explore features
type VideoService interface {
	GetGallery(ctx context.Context, userID, platform string, page, limit int) ([]*model.Video, int64, error)
	GetExplore(ctx context.Context, platform string, page, limit int) ([]*model.Video, int64, error)
	TogglePublic(ctx context.Context, videoID string, userID string) (bool, error)
	GetVideo(ctx context.Context, videoID string) (*model.Video, error)
}
