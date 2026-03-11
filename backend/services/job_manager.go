package services

import (
	"aituber/models"
	"fmt"
	"sync"
	"time"
)

// JobManager handles the state of background video generation jobs
type JobManager struct {
	jobs    map[string]*models.JobStatus
	jobsMux sync.RWMutex
}

// NewJobManager creates a new instance of job manager
func NewJobManager() *JobManager {
	return &JobManager{
		jobs: make(map[string]*models.JobStatus),
	}
}

// CreateJob creates a new job in memory
func (jm *JobManager) CreateJob(jobID, platform, contentName string) *models.JobStatus {
	job := &models.JobStatus{
		JobID:       jobID,
		Platform:    platform,
		ContentName: contentName,
		Status:      "processing",
		Progress:    0,
		CurrentStep: "Initializing",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	jm.jobsMux.Lock()
	jm.jobs[jobID] = job
	jm.jobsMux.Unlock()

	return job
}

// GetJob retrieves a job status thread-safely
func (jm *JobManager) GetJob(jobID string) (*models.JobStatus, bool) {
	jm.jobsMux.RLock()
	defer jm.jobsMux.RUnlock()
	job, exists := jm.jobs[jobID]
	// Provide a copy of the job mapping to avoid unwanted mutations.
	// For now, return direct reference as properties are value types/strings,
	// but caller must not mutate directly instead use manager methods.
	return job, exists
}

// UpdateProgress updates job's progress and current step
func (jm *JobManager) UpdateProgress(jobID string, step string, progress int) error {
	jm.jobsMux.Lock()
	defer jm.jobsMux.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	job.CurrentStep = step
	job.Progress = progress
	job.UpdatedAt = time.Now()

	return nil
}

// MarkFailed marks a job as failed
func (jm *JobManager) MarkFailed(jobID string, err error) error {
	jm.jobsMux.Lock()
	defer jm.jobsMux.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	job.Status = "failed"
	job.Error = err
	job.UpdatedAt = time.Now()

	return nil
}

// MarkCompleted marks a job as successfully generated
func (jm *JobManager) MarkCompleted(jobID, videoPath, savedPath string) error {
	jm.jobsMux.Lock()
	defer jm.jobsMux.Unlock()

	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	job.Status = "completed"
	job.Progress = 100
	job.CurrentStep = "Complete"
	job.VideoPath = videoPath
	job.SavedPath = savedPath
	job.UpdatedAt = time.Now()

	return nil
}
