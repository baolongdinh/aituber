package handler

import (
	"aituber/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VoiceHandler struct {
	voiceCatalog *service.VoiceCatalog
}

func NewVoiceHandler(voiceCatalog *service.VoiceCatalog) *VoiceHandler {
	return &VoiceHandler{
		voiceCatalog: voiceCatalog,
	}
}

// GetVoices returns available voices filtered by provider and optionally gender
func (h *VoiceHandler) GetVoices(c *gin.Context) {
	provider := c.DefaultQuery("provider", "fpt")
	gender := c.Query("gender") // Optional filter

	// Validate provider
	if !h.voiceCatalog.IsSupportedProvider(provider) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid provider",
			"message": "provider must be 'fpt' or 'hub'",
		})
		return
	}

	var voices []service.VoiceMetadata
	if gender != "" {
		voices = h.voiceCatalog.GetVoicesByGender(provider, gender)
	} else {
		voices = h.voiceCatalog.GetVoicesByProvider(provider)
	}

	c.JSON(http.StatusOK, gin.H{
		"provider": provider,
		"gender_filter": gender,
		"voices": voices,
	})
}

// GetVoiceCatalog returns the complete voice catalog for frontend
func (h *VoiceHandler) GetVoiceCatalog(c *gin.Context) {
	allVoices := h.voiceCatalog.GetAllVoices()
	
	// Group by provider for easier frontend consumption
	response := gin.H{
		"providers": gin.H{
			"fpt": h.voiceCatalog.GetVoicesByProvider("fpt"),
			"hub": h.voiceCatalog.GetVoicesByProvider("hub"),
		},
		"all_voices": allVoices,
	}

	c.JSON(http.StatusOK, response)
}
