package service

import (
	"aituber/internal/model"
	"aituber/pkg/jwtutil"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock of the UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByWalletAddress(ctx context.Context, address string) (*model.User, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateNonce(ctx context.Context, id, nonce string) error {
	args := m.Called(ctx, id, nonce)
	return args.Error(0)
}

func (m *MockUserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func TestAuthService_GetNonce(t *testing.T) {
	repo := new(MockUserRepository)
	jwt := jwtutil.NewManager("secret", 1)
	svc := NewAuthService(repo, jwt)

	wallet := "0x123"
	user := &model.User{
		BaseModel:     model.BaseModel{ID: "user-123"},
		WalletAddress: wallet,
		Nonce:         "old-nonce",
	}

	// First call to GetOrCreateUser (inside GetNonce)
	repo.On("FindByWalletAddress", mock.Anything, wallet).Return(user, nil)
	// Rotate nonce
	repo.On("UpdateNonce", mock.Anything, "user-123", mock.Anything).Return(nil)

	nonce, appErr := svc.GetNonce(context.Background(), wallet)
	assert.Nil(t, appErr)
	assert.NotEmpty(t, nonce)
	assert.Len(t, nonce, 32) // SIWE nonce hex length
	repo.AssertExpectations(t)
}
