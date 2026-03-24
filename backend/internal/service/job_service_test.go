package service

import (
	"aituber/internal/model"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockJobRepository struct {
	mock.Mock
}

func (m *MockJobRepository) Create(ctx context.Context, job *model.Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockJobRepository) FindByID(ctx context.Context, id string) (*model.Job, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Job), args.Error(1)
}

func (m *MockJobRepository) FindByUserID(ctx context.Context, userID, platform string, page, limit int) ([]*model.Job, int64, error) {
	args := m.Called(ctx, userID, platform, page, limit)
	return args.Get(0).([]*model.Job), args.Get(1).(int64), args.Error(2)
}

func (m *MockJobRepository) FindActiveByUserID(ctx context.Context, userID, platform string) (*model.Job, error) {
	args := m.Called(ctx, userID, platform)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Job), args.Error(1)
}

func (m *MockJobRepository) UpdateStatus(ctx context.Context, id, status, currentStep string, progress int) error {
	args := m.Called(ctx, id, status, currentStep, progress)
	return args.Error(0)
}

func (m *MockJobRepository) UpdateOutput(ctx context.Context, id, videoPath, savedPath, thumbnailPath string) error {
	args := m.Called(ctx, id, videoPath, savedPath, thumbnailPath)
	return args.Error(0)
}

func (m *MockJobRepository) UpdateError(ctx context.Context, id, errMsg string) error {
	args := m.Called(ctx, id, errMsg)
	return args.Error(0)
}

func (m *MockJobRepository) UpdateTitle(ctx context.Context, id, title string) error {
	args := m.Called(ctx, id, title)
	return args.Error(0)
}

func (m *MockJobRepository) UpdateCheckpoint(ctx context.Context, id string, data []byte) error {
	args := m.Called(ctx, id, data)
	return args.Error(0)
}

type MockSeriesRepository struct {
	mock.Mock
}

func (m *MockSeriesRepository) Create(ctx context.Context, series *model.Series) error {
	args := m.Called(ctx, series)
	return args.Error(0)
}

func (m *MockSeriesRepository) FindByID(ctx context.Context, id string) (*model.Series, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Series), args.Error(1)
}

func (m *MockSeriesRepository) FindActiveByUserID(ctx context.Context, userID, platform string) (*model.Series, error) {
	args := m.Called(ctx, userID, platform)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Series), args.Error(1)
}

func (m *MockSeriesRepository) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockVideoRepository struct {
	mock.Mock
}

func (m *MockVideoRepository) Create(ctx context.Context, video *model.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *MockVideoRepository) FindByID(ctx context.Context, id string) (*model.Video, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Video), args.Error(1)
}

func (m *MockVideoRepository) FindByUserID(ctx context.Context, userID, platform string, page, limit int) ([]*model.Video, int64, error) {
	args := m.Called(ctx, userID, platform, page, limit)
	return args.Get(0).([]*model.Video), args.Get(1).(int64), args.Error(2)
}

func (m *MockVideoRepository) FindPublic(ctx context.Context, platform string, page, limit int) ([]*model.Video, int64, error) {
	args := m.Called(ctx, platform, page, limit)
	return args.Get(0).([]*model.Video), args.Get(1).(int64), args.Error(2)
}

func (m *MockVideoRepository) SetPublic(ctx context.Context, id string, isPublic bool) error {
	args := m.Called(ctx, id, isPublic)
	return args.Error(0)
}

func (m *MockVideoRepository) IncrementView(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestJobService_CreateJob(t *testing.T) {
	jobRepo := new(MockJobRepository)
	seriesRepo := new(MockSeriesRepository)
	videoRepo := new(MockVideoRepository)
	svc := NewJobService(jobRepo, seriesRepo, videoRepo)

	jobRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Job")).Return(nil)

	job, err := svc.CreateJob(context.Background(), "user-1", "youtube", "My Video", "Topic", "Voice", "tts")
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "user-1", job.UserID)
	jobRepo.AssertExpectations(t)
}

func TestJobService_UpdateProgress(t *testing.T) {
	jobRepo := new(MockJobRepository)
	seriesRepo := new(MockSeriesRepository)
	videoRepo := new(MockVideoRepository)
	svc := NewJobService(jobRepo, seriesRepo, videoRepo)

	jobRepo.On("UpdateStatus", mock.Anything, "job-1", "processing", "Scripting", 10).Return(nil)

	err := svc.UpdateProgress(context.Background(), "job-1", "Scripting", 10)
	assert.NoError(t, err)
	jobRepo.AssertExpectations(t)
}
