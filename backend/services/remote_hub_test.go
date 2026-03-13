package services

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestRemoteHub_ActualCall(t *testing.T) {
	// Skip if not running manual integration tests
	if os.Getenv("RUN_REMOTE_TEST") != "true" {
		t.Skip("Skipping remote hub actual call test. Set RUN_REMOTE_TEST=true to run.")
	}

	remoteURL := "http://10.0.0.224:8081"
	remoteToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJsdWFuYWlveiJ9.E6aKblsqO9SovQAv4wDgD259iFYTSvp3xHjJoZHuUEo"

	prompt := "Urban alleyway at dusk with cinematic lighting. A tall, statuesque Vietnamese high-fashion model strides elegantly in a mid-distance full-body shot from an angular perspective, bold editorial composition with strong contrast and tactile materials. She wears a rose-gold metallic trench coat with deconstructed elements over a black long-sleeved textured turtleneck, paired with forest-green pleated pants with raw hems and soft fabric texture. She has long dark braided hair, refined Vietnamese facial features and a smooth medium complexion, carrying a vibrant yellow designer handbag with geometric details and a structured silhouette, along with white architectural sneakers featuring bold geometric cutouts. Urban grit meets high-fashion impact with dramatic dusk reflections on alley surfaces, extreme clarity and layered materials, ultra-smooth transparent high-definition film look with no noise, no grain, no blur, no vintage effect, rendered with ultra-sharp photorealistic detail like a brand-new professional fashion photograph"

	// Initialize StockVideoService with minimal requirements
	sv := NewStockVideoService(
		"",        // Pexels API Key (not needed for this test)
		"./temp",  // Temp Dir
		"./cache", // Cache Dir
		nil,       // Gemini Service
		nil,       // HF Service
		"",        // Local Hub URL
		remoteURL,
		remoteToken,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	fmt.Printf("Starting remote hub test call...\n")
	fmt.Printf("URL: %s\n", remoteURL)

	start := time.Now()
	imgBytes, err := sv.generateImageRemoteHub(ctx, prompt, "landscape")
	if err != nil {
		t.Fatalf("Remote Hub call failed: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("Success! Received %d bytes of image data in %v\n", len(imgBytes), duration)

	// Optionally save the resulting image for manual inspection
	outputPath := "remote_test_output.png"
	err = os.WriteFile(outputPath, imgBytes, 0644)
	if err != nil {
		t.Errorf("Failed to save output image: %v", err)
	} else {
		fmt.Printf("Saved output image to: %s\n", outputPath)
	}
}
