package utils

import (
	"strings"
	"testing"
)

func TestBurnSubtitlesCommand(t *testing.T) {
	// Mock RunFFmpegFunc to capture arguments
	var capturedArgs []string
	originalRunFFmpeg := RunFFmpegFunc
	RunFFmpegFunc = func(args []string) error {
		capturedArgs = args
		return nil
	}
	defer func() { RunFFmpegFunc = originalRunFFmpeg }()

	tests := []struct {
		name         string
		orientation  string
		expectMargin string
	}{
		{
			name:         "TikTok Portrait Style",
			orientation:  "portrait",
			expectMargin: "MarginV=80",
		},
		{
			name:         "YouTube Landscape Style",
			orientation:  "landscape",
			expectMargin: "MarginV=50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedArgs = nil
			err := BurnSubtitles("input.mp4", "subs.srt", "output.mp4", tt.orientation)
			if err != nil {
				t.Fatalf("BurnSubtitles failed: %v", err)
			}

			// Find the -vf argument
			vfArg := ""
			for i, arg := range capturedArgs {
				if arg == "-vf" && i+1 < len(capturedArgs) {
					vfArg = capturedArgs[i+1]
					break
				}
			}

			if vfArg == "" {
				t.Fatal("FFmpeg command missing -vf argument")
			}

			if !strings.Contains(vfArg, tt.expectMargin) {
				t.Errorf("Expected -vf to contain %s, but got: %s", tt.expectMargin, vfArg)
			}
		})
	}
}
