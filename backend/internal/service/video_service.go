package service

import (
	"aituber/internal/model"
	"aituber/internal/repository"
	"context"
	"fmt"
)

type videoServiceImpl struct {
	videoRepo repository.VideoRepository
}

func NewVideoService(videoRepo repository.VideoRepository) VideoService {
	return &videoServiceImpl{videoRepo: videoRepo}
}

func (s *videoServiceImpl) GetGallery(ctx context.Context, userID, platform string, page, limit int) ([]*model.Video, int64, error) {
	return s.videoRepo.FindByUserID(ctx, userID, platform, page, limit)
}

func (s *videoServiceImpl) GetExplore(ctx context.Context, platform string, page, limit int) ([]*model.Video, int64, error) {
	return s.videoRepo.FindPublic(ctx, platform, page, limit)
}

func (s *videoServiceImpl) TogglePublic(ctx context.Context, videoID string, userID string) (bool, error) {
	video, err := s.videoRepo.FindByID(ctx, videoID)
	if err != nil {
		return false, err
	}
	if video == nil {
		return false, fmt.Errorf("video not found")
	}
	if video.UserID != userID {
		return false, fmt.Errorf("unauthorized to share this video")
	}

	newStatus := !video.IsPublic
	if err := s.videoRepo.SetPublic(ctx, videoID, newStatus); err != nil {
		return false, err
	}
	return newStatus, nil
}

func (s *videoServiceImpl) GetVideo(ctx context.Context, videoID string) (*model.Video, error) {
	video, err := s.videoRepo.FindByID(ctx, videoID)
	if err != nil {
		return nil, err
	}
	if video != nil && video.IsPublic {
		_ = s.videoRepo.IncrementView(ctx, videoID)
	}
	return video, nil
}
