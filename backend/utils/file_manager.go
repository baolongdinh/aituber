package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// CreateTempDir creates temporary directories for a job
func CreateTempDir(baseDir, jobID string) (string, error) {
	jobDir := filepath.Join(baseDir, jobID)

	// Create subdirectories
	dirs := []string{
		jobDir,
		filepath.Join(jobDir, "audio"),
		filepath.Join(jobDir, "video"),
		filepath.Join(jobDir, "output"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return jobDir, nil
}

// DownloadFile downloads a file from URL to destination path
func DownloadFile(url, destPath string) error {
	// Create destination directory if not exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	// Download file
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy content
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// CleanupJobFiles removes all temporary files for a job
func CleanupJobFiles(baseDir, jobID string) error {
	jobDir := filepath.Join(baseDir, jobID)
	return os.RemoveAll(jobDir)
}

// ScheduleCleanup schedules automatic cleanup after a delay
func ScheduleCleanup(baseDir, jobID string, delay time.Duration) {
	go func() {
		time.Sleep(delay)
		_ = CleanupJobFiles(baseDir, jobID)
	}()
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetFileSize returns file size in bytes
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
