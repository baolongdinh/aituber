package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetMD5Hash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "5d41402abc4b2a76b9719d911017c592"},
		{"trí tuệ nhân tạo", "8136bb7aedaf9621425f067b01531445"}, // Actual hash of UTF-8 Vietnamese
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
	}

	for _, tt := range tests {
		result := GetMD5Hash(tt.input)
		if result != tt.expected {
			t.Errorf("GetMD5Hash(%q) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestCopyFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	srcPath := filepath.Join(tempDir, "src.txt")
	dstPath := filepath.Join(tempDir, "dst.txt")
	content := []byte("hello world")

	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	if err := CopyFile(srcPath, dstPath); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(content) {
		t.Errorf("Copied content mismatch. Got %q, want %q", string(got), string(content))
	}
}
