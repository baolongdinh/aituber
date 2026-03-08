package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// HuggingFaceService handles image generation via HuggingFace Inference API
type HuggingFaceService struct {
	token      string
	httpClient *http.Client
}

// NewHuggingFaceService creates a new HuggingFace service
func NewHuggingFaceService(token string) *HuggingFaceService {
	return &HuggingFaceService{
		token: token,
		httpClient: &http.Client{
			Timeout: 3 * time.Minute, // FLUX can take a while cold-starting
		},
	}
}

// HasToken returns true if an HF token is configured
func (hf *HuggingFaceService) HasToken() bool {
	return hf.token != ""
}

// GenerateImageForKeyword calls FLUX.1-schnell on HuggingFace to generate a B-roll image.
// Returns raw PNG/JPEG bytes ready to be saved and animated by FFmpeg.
// orientation: "portrait" or "landscape" – controls the prompt hint (HF doesn't support aspect ratio natively).
func (hf *HuggingFaceService) GenerateImageForKeyword(keyword, orientation string) ([]byte, error) {
	if !hf.HasToken() {
		return nil, fmt.Errorf("HuggingFace token not configured")
	}

	// Craft a cinematic B-roll prompt
	orientHint := "wide cinematic landscape composition, 16:9"
	if orientation == "portrait" {
		orientHint = "vertical phone portrait composition, 9:16, tall"
	}

	prompt := fmt.Sprintf(
		"Professional cinematic B-roll stock photo: %s. "+
			"%s, dramatic lighting, shallow depth of field, photorealistic, "+
			"no text, no watermarks, high resolution.",
		keyword, orientHint,
	)

	// FLUX.1-schnell via HuggingFace Inference API
	// Returns binary image data directly (JPEG/PNG)
	apiURL := "https://router.huggingface.co/hf-inference/models/black-forest-labs/FLUX.1-schnell"

	reqBody := map[string]interface{}{
		"inputs": prompt,
		"parameters": map[string]interface{}{
			"num_inference_steps": 4, // schnell: 4 steps is optimal
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal HF request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HF request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+hf.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "image/png")

	log.Printf("[HuggingFace] Generating %s image for: %q", orientation, keyword)

	resp, err := hf.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HF request failed: %w", err)
	}
	defer resp.Body.Close()

	// HF API returns 503 when the model is loading – worth retrying
	if resp.StatusCode == http.StatusServiceUnavailable {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HF model loading (503) – retry later: %s", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HF API returned status %d: %s", resp.StatusCode, string(body))
	}

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HF image response: %w", err)
	}

	if len(imgBytes) < 1024 {
		return nil, fmt.Errorf("HF returned suspiciously small image (%d bytes)", len(imgBytes))
	}

	log.Printf("[HuggingFace] Generated image for %q (%d bytes)", keyword, len(imgBytes))
	return imgBytes, nil
}
