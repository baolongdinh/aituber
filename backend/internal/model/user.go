package model

// User represents an authenticated user (Wallet-based auth via SIWE)
type User struct {
	BaseModel
	WalletAddress string `gorm:"uniqueIndex;not null" json:"wallet_address"` // EVM wallet address (lowercase)
	Name          string `json:"name"`
	AvatarURL     string `json:"avatar_url"`
	Nonce         string `gorm:"not null;default:''" json:"-"` // SIWE nonce, rotated after each login
}

func (User) TableName() string { return "users" }
