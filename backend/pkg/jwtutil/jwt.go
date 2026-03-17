package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the JWT payload
type Claims struct {
	UserID    string `json:"user_id"`
	UserEmail string `json:"user_email"`
	jwt.RegisteredClaims
}

// Manager handles JWT operations
type Manager struct {
	secret      []byte
	expiryHours int
}

// NewManager creates a new JWT Manager
func NewManager(secret string, expiryHours int) *Manager {
	return &Manager{
		secret:      []byte(secret),
		expiryHours: expiryHours,
	}
}

// Generate creates a signed JWT token for a user
func (m *Manager) Generate(userID, userEmail string) (string, error) {
	claims := Claims{
		UserID:    userID,
		UserEmail: userEmail,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(m.expiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Verify parses and validates a JWT token, returning Claims
func (m *Manager) Verify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
