package handlers

import (
	"aituber/utils"
	"os"
	"path/filepath"
	"testing"
)

func TestVideoHandler_BuildFinalConcatList(t *testing.T) {
	// Create a temporary directory to host mock static files
	tmpDir, err := os.MkdirTemp("", "static_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create dummy intro and outro files
	introPath := filepath.Join(tmpDir, "intro_video.mp4")
	outroPath := filepath.Join(tmpDir, "outro_video.mp4")

	if err := os.WriteFile(introPath, []byte("dummy intro"), 0644); err != nil {
		t.Fatalf("Failed to write mock intro file: %v", err)
	}
	if err := os.WriteFile(outroPath, []byte("dummy outro"), 0644); err != nil {
		t.Fatalf("Failed to write mock outro file: %v", err)
	}

	mainVideoPath := "/tmp/mock_main_video.mp4"

	// No need for VideoHandler instance for this utility test

	t.Run("YouTube Platform - Includes Intro and Outro", func(t *testing.T) {
		concatList := utils.BuildFinalConcatList("youtube", introPath, outroPath, mainVideoPath)

		if len(concatList) != 3 {
			t.Errorf("Expected concat list length to be 3 (intro, main, outro), got %d", len(concatList))
		}

		if concatList[0] != introPath {
			t.Errorf("Expected first item to be introPath: %s, got: %s", introPath, concatList[0])
		}
		if concatList[1] != mainVideoPath {
			t.Errorf("Expected second item to be mainVideoPath: %s, got: %s", mainVideoPath, concatList[1])
		}
		if concatList[2] != outroPath {
			t.Errorf("Expected third item to be outroPath: %s, got: %s", outroPath, concatList[2])
		}
	})

	t.Run("TikTok Platform - Excludes Intro and Outro", func(t *testing.T) {
		concatList := utils.BuildFinalConcatList("tiktok", introPath, outroPath, mainVideoPath)

		if len(concatList) != 1 {
			t.Errorf("Expected concat list length to be 1 (main video only for tiktok), got %d", len(concatList))
		}

		if concatList[0] != mainVideoPath {
			t.Errorf("Expected single item to be mainVideoPath: %s, got: %s", mainVideoPath, concatList[0])
		}
	})

	t.Run("YouTube Platform - Missing static files gracefully handled", func(t *testing.T) {
		nonExistentIntro := filepath.Join(tmpDir, "does_not_exist_intro.mp4")
		nonExistentOutro := filepath.Join(tmpDir, "does_not_exist_outro.mp4")

		concatList := utils.BuildFinalConcatList("youtube", nonExistentIntro, nonExistentOutro, mainVideoPath)

		if len(concatList) != 1 {
			t.Errorf("Expected concat list length to be 1 when static files are missing, got %d", len(concatList))
		}
		if concatList[0] != mainVideoPath {
			t.Errorf("Expected single item to be mainVideoPath: %s, got: %s", mainVideoPath, concatList[0])
		}
	})
}

func TestVideoHandler_Dummy(t *testing.T) {
	// Placeholder to keep the file if needed, or we could delete it if empty.
	// For now, let's just remove the broken part.
}
