package repository

import (
	"context"
	"errors"

	"aituber/internal/model"

	"gorm.io/gorm"
)

type videoRepository struct{ db *gorm.DB }

// NewVideoRepository creates a GORM-backed VideoRepository
func NewVideoRepository(db *gorm.DB) VideoRepository {
	return &videoRepository{db: db}
}

func (r *videoRepository) Create(ctx context.Context, video *model.Video) error {
	return r.db.WithContext(ctx).Create(video).Error
}

func (r *videoRepository) FindByID(ctx context.Context, id string) (*model.Video, error) {
	var video model.Video
	err := r.db.WithContext(ctx).First(&video, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &video, err
}

func (r *videoRepository) FindByUserID(ctx context.Context, userID, platform string, page, limit int) ([]*model.Video, int64, error) {
	var videos []*model.Video
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Video{}).Where("user_id = ?", userID)
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&videos).Error
	return videos, total, err
}

func (r *videoRepository) FindPublic(ctx context.Context, platform string, page, limit int) ([]*model.Video, int64, error) {
	var videos []*model.Video
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Video{}).Where("is_public = true")
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("view_count DESC, created_at DESC").Offset(offset).Limit(limit).Find(&videos).Error
	return videos, total, err
}

func (r *videoRepository) SetPublic(ctx context.Context, id string, isPublic bool) error {
	return r.db.WithContext(ctx).Model(&model.Video{}).
		Where("id = ?", id).
		Update("is_public", isPublic).Error
}

func (r *videoRepository) IncrementView(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&model.Video{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}
