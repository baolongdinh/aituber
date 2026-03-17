package repository

import (
	"context"
	"errors"

	"aituber/internal/model"

	"gorm.io/gorm"
)

type jobRepository struct{ db *gorm.DB }

// NewJobRepository creates a GORM-backed JobRepository
func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepository{db: db}
}

func (r *jobRepository) Create(ctx context.Context, job *model.Job) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *jobRepository) FindByID(ctx context.Context, id string) (*model.Job, error) {
	var job model.Job
	err := r.db.WithContext(ctx).First(&job, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &job, err
}

func (r *jobRepository) FindByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Job, int64, error) {
	var jobs []*model.Job
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Job{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&jobs).Error
	return jobs, total, err
}

func (r *jobRepository) UpdateStatus(ctx context.Context, id, status, currentStep string, progress int) error {
	return r.db.WithContext(ctx).Model(&model.Job{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       status,
			"current_step": currentStep,
			"progress":     progress,
		}).Error
}

func (r *jobRepository) UpdateOutput(ctx context.Context, id, videoPath, savedPath string) error {
	return r.db.WithContext(ctx).Model(&model.Job{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"video_path": videoPath,
			"saved_path": savedPath,
			"status":     "completed",
			"progress":   100,
		}).Error
}

func (r *jobRepository) UpdateError(ctx context.Context, id, errMsg string) error {
	return r.db.WithContext(ctx).Model(&model.Job{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": errMsg,
		}).Error
}

// SeriesRepository implementation

type seriesRepository struct{ db *gorm.DB }

// NewSeriesRepository creates a GORM-backed SeriesRepository
func NewSeriesRepository(db *gorm.DB) SeriesRepository {
	return &seriesRepository{db: db}
}

func (r *seriesRepository) Create(ctx context.Context, series *model.Series) error {
	return r.db.WithContext(ctx).Create(series).Error
}

func (r *seriesRepository) FindByID(ctx context.Context, id string) (*model.Series, error) {
	var series model.Series
	err := r.db.WithContext(ctx).Preload("Jobs").First(&series, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &series, err
}

func (r *seriesRepository) UpdateStatus(ctx context.Context, id, status string) error {
	return r.db.WithContext(ctx).Model(&model.Series{}).
		Where("id = ?", id).
		Update("status", status).Error
}
