package service

import (
	"aituber/internal/model"
	"aituber/internal/repository"
	"aituber/pkg/apperror"
	"aituber/pkg/jwtutil"
	"aituber/pkg/siwe"
	"context"
	"fmt"
	"strings"
)

// AuthService handles wallet-based authentication
type AuthService interface {
	GetOrCreateUser(ctx context.Context, walletAddress string) (*model.User, *apperror.AppError)
	GetNonce(ctx context.Context, walletAddress string) (string, *apperror.AppError)
	LoginWithWallet(ctx context.Context, walletAddress, signature string) (string, *model.User, *apperror.AppError)
	UpdateProfile(ctx context.Context, userID, name string) *apperror.AppError
}

type authServiceImpl struct {
	userRepo repository.UserRepository
	jwt      *jwtutil.Manager
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo repository.UserRepository, jwt *jwtutil.Manager) AuthService {
	return &authServiceImpl{userRepo: userRepo, jwt: jwt}
}

func (s *authServiceImpl) GetOrCreateUser(ctx context.Context, walletAddress string) (*model.User, *apperror.AppError) {
	addr := strings.ToLower(walletAddress)
	user, err := s.userRepo.FindByWalletAddress(ctx, addr)
	if err != nil {
		return nil, apperror.Internal(err, "failed to find user")
	}
	if user != nil {
		return user, nil
	}

	// Auto-create user on first login
	nonce, err := siwe.GenerateNonce()
	if err != nil {
		return nil, apperror.Internal(err, "failed to generate nonce")
	}
	newUser := &model.User{
		WalletAddress: addr,
		Name:          fmt.Sprintf("User %s", addr[:6]),
		Nonce:         nonce,
	}
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, apperror.Internal(err, "failed to create user")
	}
	return newUser, nil
}

func (s *authServiceImpl) GetNonce(ctx context.Context, walletAddress string) (string, *apperror.AppError) {
	user, appErr := s.GetOrCreateUser(ctx, walletAddress)
	if appErr != nil {
		return "", appErr
	}

	// Rotate nonce each time it's requested
	nonce, err := siwe.GenerateNonce()
	if err != nil {
		return "", apperror.Internal(err, "failed to generate nonce")
	}
	if err := s.userRepo.UpdateNonce(ctx, user.ID, nonce); err != nil {
		return "", apperror.Internal(err, "failed to update nonce")
	}
	return nonce, nil
}

func (s *authServiceImpl) LoginWithWallet(ctx context.Context, walletAddress, signature string) (string, *model.User, *apperror.AppError) {
	addr := strings.ToLower(walletAddress)
	user, err := s.userRepo.FindByWalletAddress(ctx, addr)
	if err != nil {
		return "", nil, apperror.Internal(err, "failed to find user")
	}
	if user == nil {
		return "", nil, apperror.NotFound("user not found — request a nonce first")
	}

	// Rebuild the exact message that was signed on the frontend
	message := siwe.BuildMessage(addr, user.Nonce)

	// Verify signature
	if err := siwe.VerifySignature(addr, message, signature); err != nil {
		return "", nil, apperror.Unauthorized("invalid signature: " + err.Error())
	}

	// Rotate nonce after successful login (one-time use)
	newNonce, _ := siwe.GenerateNonce()
	_ = s.userRepo.UpdateNonce(ctx, user.ID, newNonce)

	// Issue JWT
	token, err := s.jwt.Generate(user.ID, user.WalletAddress)
	if err != nil {
		return "", nil, apperror.Internal(err, "failed to generate token")
	}
	return token, user, nil
}

func (s *authServiceImpl) UpdateProfile(ctx context.Context, userID, name string) *apperror.AppError {
	name = strings.TrimSpace(name)
	if name == "" {
		return apperror.BadRequest("name cannot be empty")
	}

	// Check for unique name
	existing, err := s.userRepo.FindByName(ctx, name)
	if err != nil {
		return apperror.Internal(err, "failed to check name uniqueness")
	}
	if existing != nil && existing.ID != userID {
		return apperror.BadRequest("name already taken")
	}

	// Fetch user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperror.Internal(err, "failed to fetch user")
	}
	if user == nil {
		return apperror.NotFound("user not found")
	}

	// Update name
	user.Name = name
	if err := s.userRepo.Update(ctx, user); err != nil {
		return apperror.Internal(err, "failed to update user")
	}

	return nil
}
