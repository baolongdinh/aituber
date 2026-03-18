package repository

import (
	"context"
	"errors"

	"aituber/internal/model"

	"gorm.io/gorm"
)

type userRepository struct{ db *gorm.DB }

// NewUserRepository creates a GORM-backed UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) FindByWalletAddress(ctx context.Context, address string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, "wallet_address = ?", address).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) UpdateNonce(ctx context.Context, id, nonce string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", id).
		Update("nonce", nonce).Error
}

func (r *userRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, "name = ?", name).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}
