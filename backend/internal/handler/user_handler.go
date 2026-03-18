package handler

import (
	"aituber/internal/dto"
	"aituber/internal/repository"
	"aituber/internal/service"
	"aituber/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo repository.UserRepository
	authSvc  service.AuthService
}

func NewUserHandler(userRepo repository.UserRepository, authSvc service.AuthService) *UserHandler {
	return &UserHandler{userRepo: userRepo, authSvc: authSvc}
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

// UpdateProfile godoc
// @Summary Update current user profile (name)
// @Tags User
// @Security BearerAuth
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	appErr := h.authSvc.UpdateProfile(c.Request.Context(), userID, req.Name)
	if appErr != nil {
		response.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	response.OK(c, gin.H{"message": "profile updated"})
}
