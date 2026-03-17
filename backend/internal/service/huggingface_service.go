package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

// HuggingFaceService handles image generation via HuggingFace Inference API.
// Supports multiple tokens in round-robin to spread rate-limit pressure.
type HuggingFaceService struct {
	tokens     []string
	counter    atomic.Uint64 // round-robin counter
	httpClient *http.Client
}

// hfModels is the ordered list of text-to-image models to try.
// Ordered fast → highest quality. Each model is retried up to hfMaxRetriesPerModel times before falling through to the next.
var hfModels = []string{
	"ByteDance/SDXL-Lightning",                    // fast, 8 steps
	"Tongyi-MAI/Z-Image-Turbo",                    // fast, 8 steps
	"stable-diffusion-v1-5/stable-diffusion-v1-5", // very stable fallback
	"stabilityai/stable-diffusion-xl-base-1.0",    // good quality
	"stabilityai/stable-diffusion-3.5-large",      // high quality
	"black-forest-labs/FLUX.1-dev",                // highest quality (slower)
	"black-forest-labs/FLUX.1-schnell",            // fast FLUX fallback
}

// t2vModels is the ordered list of text-to-video models to try.
// Ordered by quality and recency as requested.
var t2vModels = []string{
	"tencent/HunyuanVideo-1.5", // top-tier, very recent
	"Wan-AI/Wan2.2-T2V-A14B",   // high quality 14B
	"genmo/mochi-1-preview",    // very popular, high quality
	"Wan-AI/Wan2.2-TI2V-5B",    // high quality, slightly smaller
	"zai-org/CogVideoX-5b",     // stable fallback
	"Wan-AI/Wan2.1-T2V-14B",    // older but very reliable
}

const hfMaxRetriesPerModel = 3

// NewHuggingFaceService creates a new HuggingFace service.
// tokens is a slice of HF API tokens that will be used in round-robin.
func NewHuggingFaceService(tokens []string) *HuggingFaceService {
	return &HuggingFaceService{
		tokens: tokens,
		httpClient: &http.Client{
			Timeout: 3 * time.Minute, // models can take a while cold-starting
		},
	}
}

// HasToken returns true if at least one HF token is configured
func (hf *HuggingFaceService) HasToken() bool {
	return len(hf.tokens) > 0
}

// nextToken returns the next token in round-robin order.
func (hf *HuggingFaceService) nextToken() string {
	idx := hf.counter.Add(1) - 1
	return hf.tokens[idx%uint64(len(hf.tokens))]
}

// GenerateImageForKeyword tries each model in hfModels (up to hfMaxRetriesPerModel retries each)
// until an image is successfully generated. Tokens are rotated in round-robin per request.
// Returns raw PNG/JPEG bytes ready to be saved and animated by FFmpeg.
// GenerateVideoForPrompt uses HuggingFace Inference Providers (e.g., fal-ai) to generate a video clip.
// provider: "fal-ai" (recommended)
// model: "genmo/mochi-1-preview" or "Wan-AI/Wan2.2-T2V-A14B"
func (hf *HuggingFaceService) GenerateVideoForPrompt(prompt, model, provider string) ([]byte, error) {
	if !hf.HasToken() {
		return nil, fmt.Errorf("HuggingFace token not configured")
	}

	// Determine models to try
	modelsToTry := t2vModels
	if model != "" {
		// If a specific model is requested, try it first, then fall back to the others
		modelsToTry = append([]string{model}, t2vModels...)
	}

	var lastErr error
	for _, currentModel := range modelsToTry {
		apiURL := fmt.Sprintf("https://router.huggingface.co/hf-inference/models/%s", currentModel)

		reqBody := map[string]interface{}{
			"inputs":   prompt,
			"provider": provider,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			lastErr = fmt.Errorf("model %s: marshal error: %w", currentModel, err)
			continue
		}

		for attempt := 1; attempt <= hfMaxRetriesPerModel; attempt++ {
			if attempt > 1 {
				backoff := time.Duration(attempt) * 5 * time.Second
				log.Printf("[HuggingFace T2V] model=%s attempt=%d/%d retrying in %s...", currentModel, attempt, hfMaxRetriesPerModel, backoff)
				time.Sleep(backoff)
			}

			token := hf.nextToken()
			log.Printf("[HuggingFace T2V] model=%s attempt=%d/%d generating video for: %q", currentModel, attempt, hfMaxRetriesPerModel, prompt)

			req, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
			if err != nil {
				lastErr = fmt.Errorf("model %s: create request: %w", currentModel, err)
				continue
			}
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("x-use-cache", "false")

			resp, err := hf.httpClient.Do(req)
			if err != nil {
				lastErr = fmt.Errorf("model %s: request failed: %w", currentModel, err)
				continue
			}

			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				lastErr = fmt.Errorf("model %s: read response: %w", currentModel, readErr)
				continue
			}

			if resp.StatusCode == http.StatusOK {
				// Some providers might return JSON with a URL instead of binary.
				if json.Valid(body) && bytes.Contains(body, []byte(`"url"`)) {
					var result struct {
						URL string `json:"url"`
					}
					if err := json.Unmarshal(body, &result); err == nil && result.URL != "" {
						log.Printf("[HuggingFace T2V] Got video URL: %s, downloading...", result.URL)
						return hf.downloadFile(result.URL)
					}
				}

				log.Printf("[HuggingFace T2V] model=%s Success (%d bytes)", currentModel, len(body))
				return body, nil
			}

			lastErr = fmt.Errorf("model %s status %d: %s", currentModel, resp.StatusCode, string(body))
			log.Printf("[HuggingFace T2V] Error: %v", lastErr)

			// Non-503/serverless-loading errors usually won't succeed on retry
			if resp.StatusCode != http.StatusServiceUnavailable && resp.StatusCode != http.StatusTooManyRequests {
				break
			}
		}
		log.Printf("[HuggingFace T2V] model=%s exhausted, trying next fallback model...", currentModel)
	}

	return nil, fmt.Errorf("all T2V models failed: %w", lastErr)
}

func (hf *HuggingFaceService) downloadFile(url string) ([]byte, error) {
	resp, err := hf.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// GenerateImageForKeyword generates a cinematic image using HuggingFace diffusion model.
// visualDesc: optional cinematic scene description from the video script (preferred over keyword when non-empty).
func (hf *HuggingFaceService) GenerateImageForKeyword(keyword, visualDesc, orientation string) ([]byte, error) {
	if !hf.HasToken() {
		return nil, fmt.Errorf("HuggingFace token not configured")
	}

	// Orientation hint to guide composition
	orientHint := "wide cinematic landscape composition, 16:9"
	if orientation == "portrait" {
		orientHint = "vertical phone portrait composition, 9:16, tall"
	}

	// Build image prompt: prefer rich visual_description from script; fall back to short keyword.
	var prompt string
	if visualDesc != "" {
		// visualDesc is a detailed cinematic description – append orientation and quality constraints.
		prompt = fmt.Sprintf(
			"%s. %s, photorealistic, high resolution, no text, no watermarks.",
			visualDesc, orientHint,
		)
	} else {
		prompt = fmt.Sprintf(
			"Professional cinematic B-roll stock photo: %s. "+
				"%s, dramatic lighting, shallow depth of field, photorealistic, "+
				"no text, no watermarks, high resolution.",
			keyword, orientHint,
		)
	}

	var lastErr error

	for _, model := range hfModels {
		apiURL := fmt.Sprintf("https://router.huggingface.co/hf-inference/models/%s", model)

		// Choose inference steps based on model family
		numSteps := 20
		switch model {
		case "black-forest-labs/FLUX.1-schnell":
			numSteps = 4 // schnell is optimised for 4 steps
		case "black-forest-labs/FLUX.1-dev":
			numSteps = 25 // dev benefits from more steps
		case "ByteDance/SDXL-Lightning", "Tongyi-MAI/Z-Image-Turbo":
			numSteps = 8 // lightning/turbo models: 4-8 steps
		}

		reqBody := map[string]interface{}{
			"inputs": prompt,
			"parameters": map[string]interface{}{
				"num_inference_steps": numSteps,
			},
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			lastErr = fmt.Errorf("model %s: marshal error: %w", model, err)
			continue
		}

		for attempt := 1; attempt <= hfMaxRetriesPerModel; attempt++ {
			if attempt > 1 {
				backoff := time.Duration(attempt) * 3 * time.Second
				log.Printf("[HuggingFace] model=%s attempt=%d/%d retrying in %s...", model, attempt, hfMaxRetriesPerModel, backoff)
				time.Sleep(backoff)
			}

			// Pick next token via round-robin
			token := hf.nextToken()
			subjectLabel := keyword
			if visualDesc != "" {
				if len(visualDesc) > 80 {
					subjectLabel = visualDesc[:80] + "..."
				} else {
					subjectLabel = visualDesc
				}
			}
			log.Printf("[HuggingFace] model=%s attempt=%d/%d token=...%s generating %s image for: %q",
				model, attempt, hfMaxRetriesPerModel, token[max(0, len(token)-6):], orientation, subjectLabel)

			req, err := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
			if err != nil {
				lastErr = fmt.Errorf("model %s attempt %d: create request: %w", model, attempt, err)
				continue
			}
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "image/png")

			resp, err := hf.httpClient.Do(req)
			if err != nil {
				lastErr = fmt.Errorf("model %s attempt %d: request failed: %w", model, attempt, err)
				continue
			}

			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				lastErr = fmt.Errorf("model %s attempt %d: read response: %w", model, attempt, readErr)
				continue
			}

			switch resp.StatusCode {
			case http.StatusOK:
				if len(body) < 1024 {
					lastErr = fmt.Errorf("model %s attempt %d: suspiciously small image (%d bytes)", model, attempt, len(body))
					continue
				}
				log.Printf("[HuggingFace] model=%s succeeded (%d bytes)", model, len(body))
				return body, nil

			case http.StatusServiceUnavailable:
				// Model is still loading – worth retrying
				lastErr = fmt.Errorf("model %s attempt %d: model loading (503): %s", model, attempt, string(body))
				continue

			default:
				lastErr = fmt.Errorf("model %s attempt %d: status %d: %s", model, attempt, resp.StatusCode, string(body))
				// Non-503 errors are unlikely to recover on retry – move to next model
				goto nextModel
			}
		}

	nextModel:
		log.Printf("[HuggingFace] model=%s exhausted all retries, trying next model...", model)
	}

	return nil, fmt.Errorf("all HuggingFace models failed: %w", lastErr)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
