package handler

import (
	"aituber/internal/dto"
	"aituber/internal/repository"
	"aituber/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// GetMe godoc
// @Summary Get current user profile
// @Tags User
// @Security BearerAuth
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch user profile")
		return
	}
	if user == nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}

	response.OK(c, dto.UserResponse{
		ID:            user.ID,
		WalletAddress: user.WalletAddress,
		Name:          user.Name,
		AvatarURL:     user.AvatarURL,
	})
}
