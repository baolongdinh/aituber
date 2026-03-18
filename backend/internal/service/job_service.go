package service

import (
	"aituber/internal/model"
	"aituber/internal/repository"
	"context"
	"fmt"
	"sync"
)

type jobServiceImpl struct {
	jobRepo    repository.JobRepository
	seriesRepo repository.SeriesRepository
	videoRepo  repository.VideoRepository

	// In-memory cache for live status (similar to old JobManager)
	// We still keep this for high-frequency polling/updates before final DB sync if needed,
	// but for now we'll sync directly to DB to ensure persistence across restarts.
	mu   sync.RWMutex
	live map[string]*model.Job
}

func NewJobService(jobRepo repository.JobRepository, seriesRepo repository.SeriesRepository, videoRepo repository.VideoRepository) JobService {
	return &jobServiceImpl{
		jobRepo:    jobRepo,
		seriesRepo: seriesRepo,
		videoRepo:  videoRepo,
		live:       make(map[string]*model.Job),
	}
}

func (s *jobServiceImpl) CreateJob(ctx context.Context, userID, platform, contentName, topic, voice, ttsProvider string) (*model.Job, error) {
	job := &model.Job{
		UserID:      userID,
		Platform:    platform,
		ContentName: contentName,
		Topic:       topic,
		Voice:       voice,
		TTSProvider: ttsProvider,
		Status:      "processing",
		Progress:    0,
		CurrentStep: "Initializing",
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	s.mu.Lock()
	s.live[job.ID] = job
	s.mu.Unlock()

	return job, nil
}

func (s *jobServiceImpl) GetJob(ctx context.Context, jobID string) (*model.Job, error) {
	// Check cache first
	s.mu.RLock()
	if job, ok := s.live[jobID]; ok {
		s.mu.RUnlock()
		return job, nil
	}
	s.mu.RUnlock()

	// Fallback to DB
	return s.jobRepo.FindByID(ctx, jobID)
}

func (s *jobServiceImpl) ListUserJobs(ctx context.Context, userID, platform string, page, limit int) ([]*model.Job, int64, error) {
	return s.jobRepo.FindByUserID(ctx, userID, platform, page, limit)
}

func (s *jobServiceImpl) UpdateProgress(ctx context.Context, jobID, step string, progress int) error {
	// Sync to DB
	if err := s.jobRepo.UpdateStatus(ctx, jobID, "processing", step, progress); err != nil {
		return err
	}

	// Update cache
	s.mu.Lock()
	if job, ok := s.live[jobID]; ok {
		job.CurrentStep = step
		job.Progress = progress
		job.Status = "processing"
	}
	s.mu.Unlock()

	return nil
}

func (s *jobServiceImpl) MarkFailed(ctx context.Context, jobID string, err error) error {
	errMsg := err.Error()
	if err := s.jobRepo.UpdateError(ctx, jobID, errMsg); err != nil {
		return err
	}

	s.mu.Lock()
	if job, ok := s.live[jobID]; ok {
		job.Status = "failed"
		job.ErrorMsg = &errMsg
	}
	s.mu.Unlock()

	return nil
}

func (s *jobServiceImpl) MarkCompleted(ctx context.Context, jobID, videoPath, savedPath, thumbnailPath string) error {
	job, err := s.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	// Create Video record
	video := &model.Video{
		UserID:       job.UserID,
		JobID:        jobID,
		Title:        job.ContentName,
		Platform:     job.Platform,
		ContentType:  job.Type,
		FilePath:     savedPath,
		ThumbnailURL: thumbnailPath,
		DurationSec:  0, // TODO: Extract duration
	}
	if err := s.videoRepo.Create(ctx, video); err != nil {
		return fmt.Errorf("failed to create video record: %w", err)
	}

	s.mu.Lock()
	if job, ok := s.live[jobID]; ok {
		job.Status = "completed"
		job.Progress = 100
		job.CurrentStep = "Completed"
		job.VideoPath = videoPath
		job.SavedPath = savedPath
		job.ThumbnailURL = thumbnailPath
	}
	s.mu.Unlock()

	return s.jobRepo.UpdateOutput(ctx, jobID, videoPath, savedPath, thumbnailPath)
}

func (s *jobServiceImpl) CreateSeries(ctx context.Context, userID, topic, platform, contentName string, numParts int) (*model.Series, error) {
	series := &model.Series{
		UserID:      userID,
		Topic:       topic,
		Platform:    platform,
		ContentName: contentName,
		NumParts:    numParts,
		Status:      "processing",
	}
	if err := s.seriesRepo.Create(ctx, series); err != nil {
		return nil, err
	}
	return series, nil
}

func (s *jobServiceImpl) GetSeries(ctx context.Context, seriesID string) (*model.Series, error) {
	return s.seriesRepo.FindByID(ctx, seriesID)
}

func (s *jobServiceImpl) UpdateSeriesStatus(ctx context.Context, seriesID, status string) error {
	return s.seriesRepo.UpdateStatus(ctx, seriesID, status)
}

func (s *jobServiceImpl) CreateSeriesPartJob(ctx context.Context, userID, seriesID string, partIndex int, platform, contentName, topic, voice, ttsProvider string) (*model.Job, error) {
	job := &model.Job{
		UserID:      userID,
		SeriesID:    &seriesID,
		Platform:    platform,
		ContentName: contentName,
		Topic:       topic,
		Voice:       voice,
		TTSProvider: ttsProvider,
		Type:        "series_part",
		PartIndex:   partIndex,
		Status:      "queued",
		Progress:    0,
		CurrentStep: "Queued",
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create series part job: %w", err)
	}

	s.mu.Lock()
	s.live[job.ID] = job
	s.mu.Unlock()

	return job, nil
}

func (s *jobServiceImpl) GetActiveTask(ctx context.Context, userID, platform string) (*model.Job, *model.Series, error) {
	// First check for active series
	series, err := s.seriesRepo.FindActiveByUserID(ctx, userID, platform)
	if err != nil {
		return nil, nil, err
	}
	if series != nil {
		return nil, series, nil
	}

	// Then check for active standalone job
	job, err := s.jobRepo.FindActiveByUserID(ctx, userID, platform)
	if err != nil {
		return nil, nil, err
	}

	return job, nil, nil
}
func (s *jobServiceImpl) UpdateJobTitle(ctx context.Context, jobID, title string) error {
	return s.jobRepo.UpdateTitle(ctx, jobID, title)
}
