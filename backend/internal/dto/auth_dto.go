package dto

// LoginRequest is the payload for SIWE login
type LoginRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Signature     string `json:"signature" binding:"required"`
}

// TokenResponse is returned after successful login
type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

// UserResponse is a public/safe user profile
type UserResponse struct {
	ID            string `json:"id"`
	WalletAddress string `json:"wallet_address"`
	Name          string `json:"name"`
	AvatarURL     string `json:"avatar_url"`
}
