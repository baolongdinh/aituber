package repository

import (
	"aituber/internal/model"
	"context"
)

// UserRepository defines the contract for user data access
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByWalletAddress(ctx context.Context, address string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	UpdateNonce(ctx context.Context, id, nonce string) error
}

// JobRepository defines the contract for job data access
type JobRepository interface {
	Create(ctx context.Context, job *model.Job) error
	FindByID(ctx context.Context, id string) (*model.Job, error)
	FindByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Job, int64, error)
	UpdateStatus(ctx context.Context, id, status, currentStep string, progress int) error
	UpdateOutput(ctx context.Context, id, videoPath, savedPath string) error
	UpdateError(ctx context.Context, id, errMsg string) error
}

// SeriesRepository defines the contract for series data access
type SeriesRepository interface {
	Create(ctx context.Context, series *model.Series) error
	FindByID(ctx context.Context, id string) (*model.Series, error)
	UpdateStatus(ctx context.Context, id, status string) error
}

// VideoRepository defines the contract for video data access
type VideoRepository interface {
	Create(ctx context.Context, video *model.Video) error
	FindByID(ctx context.Context, id string) (*model.Video, error)
	FindByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Video, int64, error)
	FindPublic(ctx context.Context, page, limit int) ([]*model.Video, int64, error)
	SetPublic(ctx context.Context, id string, isPublic bool) error
	IncrementView(ctx context.Context, id string) error
}
