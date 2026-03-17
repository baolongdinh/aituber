package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestBurnSubtitlesReal(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real video test in short mode")
	}

	artifactDir := "/home/aiozlong/.gemini/antigravity/brain/afcf4e80-8f65-4fbc-8efc-5c0d381b0a0f"
	os.MkdirAll(artifactDir, 0755)

	inputPath := filepath.Join(artifactDir, "test_input.mp4")
	srtPath := filepath.Join(artifactDir, "test_subs.srt")
	outputPath := filepath.Join(artifactDir, "test_result_video.mp4")

	// 1. Generate 3s dummy video (Portrait 720x1280)
	t.Log("Generating dummy video...")
	cmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=3:size=720x1280:rate=30", "-c:v", "libx264", "-y", inputPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to generate dummy video: %v\nOutput: %s", err, string(out))
	}

	// 2. Create dummy SRT
	t.Log("Creating dummy SRT...")
	srtContent := "1\n00:00:00,000 --> 00:00:03,000\nThis is a test subtitle at 2/3 down.\n"
	os.WriteFile(srtPath, []byte(srtContent), 0644)

	// 3. Run BurnSubtitles
	t.Log("Running BurnSubtitles (Portrait)...")
	err := BurnSubtitles(inputPath, srtPath, outputPath, "portrait")
	if err != nil {
		t.Fatalf("BurnSubtitles failed: %v", err)
	}

	// Check file existence
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("Output video file does not exist: %v", err)
	} else {
		t.Logf("SUCCESS: Output video file created at %s", outputPath)

		// Verify with ffprobe that it has video
		probeCmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=codec_name", "-of", "default=noprint_wrappers=1:nokey=1", outputPath)
		probeOut, err := probeCmd.Output()
		if err != nil {
			t.Errorf("ffprobe failed: %v", err)
		} else {
			t.Logf("Verified output video codec: %s", string(probeOut))
		}
	}
}
