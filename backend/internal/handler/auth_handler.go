package handler

import (
	"aituber/internal/dto"
	"aituber/internal/service"
	"aituber/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc service.AuthService
}

func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// GetNonce godoc
// @Summary Request a nonce for SIWE login
// @Description Creates or rotates a nonce for the given wallet address.
// @Tags Auth
func (h *AuthHandler) GetNonce(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "address query param is required")
		return
	}

	nonce, appErr := h.authSvc.GetNonce(c.Request.Context(), address)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	response.OK(c, gin.H{"nonce": nonce})
}

// LoginWithWallet godoc
// @Summary Verify SIWE signature and login
// @Description Verifies the EIP-191 signature and issues a JWT token.
// @Tags Auth
func (h *AuthHandler) LoginWithWallet(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	token, user, appErr := h.authSvc.LoginWithWallet(c.Request.Context(), req.WalletAddress, req.Signature)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	response.OK(c, gin.H{
		"token": token,
		"user": dto.UserResponse{
			ID:            user.ID,
			WalletAddress: user.WalletAddress,
			Name:          user.Name,
			AvatarURL:     user.AvatarURL,
		},
	})
}
