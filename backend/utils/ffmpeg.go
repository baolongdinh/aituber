package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// RunFFmpegCommand executes an FFmpeg command
func RunFFmpegCommand(args []string) error {
	cmd := exec.Command("ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// GetVideoDuration returns the duration of a video file in seconds
func GetVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe error: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// GetAudioDuration returns the duration of an audio file in seconds
func GetAudioDuration(audioPath string) (float64, error) {
	return GetVideoDuration(audioPath) // Same implementation
}

// MergeAudioWithCrossfade merges audio files with crossfade effect
func MergeAudioWithCrossfade(inputFiles []string, outputFile string, crossfadeDuration float64, bitrate string) error {
	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files provided")
	}

	if len(inputFiles) == 1 {
		// Single file - just copy with normalization
		args := []string{
			"-i", inputFiles[0],
			"-af", "loudnorm",
			"-ar", "44100",
			"-ab", bitrate,
			"-y", outputFile,
		}
		return RunFFmpegCommand(args)
	}

	// Multiple files - build complex filter
	args := []string{}

	// Add input files
	for i, file := range inputFiles {
		if file == "" {
			return fmt.Errorf("empty input file path at index %d", i)
		}

		absPath, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", file, err)
		}
		args = append(args, "-i", absPath)
	}

	// Build filter complex for crossfade
	filterParts := []string{}
	lastLabel := "[0:a]"

	for i := 1; i < len(inputFiles); i++ {
		currentInput := fmt.Sprintf("[%d:a]", i)
		outputLabel := fmt.Sprintf("[a%d]", i)

		if i == len(inputFiles)-1 {
			outputLabel = "[aout]"
		}

		filter := fmt.Sprintf("%s%sacrossfade=d=%.2f:c1=tri:c2=tri%s",
			lastLabel, currentInput, crossfadeDuration, outputLabel)
		filterParts = append(filterParts, filter)

		lastLabel = outputLabel
	}

	// Add loudnorm at the end
	filterComplex := strings.Join(filterParts, ";") + ";[aout]loudnorm[final]"

	args = append(args,
		"-filter_complex", filterComplex,
		"-map", "[final]",
		"-ar", "44100",
		"-ab", bitrate,
		"-y", outputFile,
	)

	return RunFFmpegCommand(args)
}

// MergeVideosWithTransition merges video files with transition effects
func MergeVideosWithTransition(inputFiles []string, outputFile string, transitionDuration float64, fps int, resolution string) error {
	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files provided")
	}

	if len(inputFiles) == 1 {
		// Single file - just re-encode
		args := []string{
			"-i", inputFiles[0],
			"-c:v", "libx264",
			"-preset", "slow",
			"-crf", "18",
			"-r", strconv.Itoa(fps),
			"-s", resolution,
			"-y", outputFile,
		}
		return RunFFmpegCommand(args)
	}

	// Get durations to calculate offsets
	durations := make([]float64, len(inputFiles))
	for i, file := range inputFiles {
		dur, err := GetVideoDuration(file)
		if err != nil {
			return fmt.Errorf("failed to get duration of %s: %w", file, err)
		}
		durations[i] = dur
	}

	// Build filter complex
	args := []string{}

	// Add input files
	for _, file := range inputFiles {
		args = append(args, "-i", file)
	}

	// Build normalization and xfade transitions
	filterParts := []string{}

	// 1. Normalize all inputs first (resolution, fps, pixel format, sar)
	// This prevents "timebase mismatch" and "main timebase" errors in xfade
	for i := 0; i < len(inputFiles); i++ {
		// Scale to target resolution, force generic PAR, set FPS, set pixel format
		// [0:v]scale=1920:1080,setsar=1,fps=30,format=yuv420p[v0norm]
		normFilter := fmt.Sprintf("[%d:v]scale=%s,setsar=1,fps=%d,format=yuv420p[v%dnorm]",
			i, resolution, fps, i)
		filterParts = append(filterParts, normFilter)
	}

	// 2. Apply xfade transitions
	offset := 0.0
	// Start with the first normalized text
	lastLabel := "[v0norm]"

	for i := 1; i < len(inputFiles); i++ {
		offset += durations[i-1] - transitionDuration
		currentInput := fmt.Sprintf("[v%dnorm]", i)
		outputLabel := fmt.Sprintf("[v%d]", i)

		if i == len(inputFiles)-1 {
			outputLabel = "[vout]"
		}

		filter := fmt.Sprintf("%s%sxfade=transition=fade:duration=%.2f:offset=%.2f%s",
			lastLabel, currentInput, transitionDuration, offset, outputLabel)
		filterParts = append(filterParts, filter)

		lastLabel = outputLabel
	}

	filterComplex := strings.Join(filterParts, ";")

	args = append(args,
		"-filter_complex", filterComplex,
		"-map", "[vout]",
		"-c:v", "libx264",
		"-preset", "slow",
		"-crf", "18",
		"-r", strconv.Itoa(fps),
		"-y", outputFile,
	)

	return RunFFmpegCommand(args)
}

// CombineAudioVideo combines audio and video into final output
func CombineAudioVideo(videoPath, audioPath, outputPath string, videoBitrate string) error {
	args := []string{
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "libx264",
		"-preset", "slow",
		"-b:v", videoBitrate,
		"-c:a", "aac",
		"-b:a", "192k",
		"-map", "0:v:0",
		"-map", "1:a:0",
		"-shortest",
		"-y", outputPath,
	}

	return RunFFmpegCommand(args)
}

// ExtendVideo extends video duration by freezing last frame
func ExtendVideo(inputPath, outputPath string, targetDuration float64) error {
	currentDuration, err := GetVideoDuration(inputPath)
	if err != nil {
		return err
	}

	if currentDuration >= targetDuration {
		// Already long enough - just copy
		args := []string{"-i", inputPath, "-c", "copy", "-y", outputPath}
		return RunFFmpegCommand(args)
	}

	// Freeze last frame
	freezeDuration := targetDuration - currentDuration

	args := []string{
		"-i", inputPath,
		"-filter_complex",
		fmt.Sprintf("[0:v]trim=duration=%.2f,setpts=PTS-STARTPTS[v1];[0:v]trim=start=%.2f,setpts=PTS-STARTPTS,tpad=stop_duration=%.2f:stop_mode=clone[v2];[v1][v2]concat=n=2:v=1:a=0[vout]",
			currentDuration, currentDuration-0.1, freezeDuration),
		"-map", "[vout]",
		"-c:v", "libx264",
		"-preset", "slow",
		"-crf", "18",
		"-y", outputPath,
	}

	return RunFFmpegCommand(args)
}

// TrimVideo trims video to target duration
func TrimVideo(inputPath, outputPath string, targetDuration float64) error {
	args := []string{
		"-i", inputPath,
		"-t", fmt.Sprintf("%.2f", targetDuration),
		"-c", "copy",
		"-y", outputPath,
	}

	return RunFFmpegCommand(args)
}

// ConcatVideos concatenates multiple video files with audio, normalizing them
func ConcatVideos(inputFiles []string, outputPath string) error {
	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files provided")
	}

	// Build filter complex
	args := []string{}

	// Add input files
	for _, file := range inputFiles {
		args = append(args, "-i", file)
	}

	// Filter complex for normalization and concat
	filterParts := []string{}

	for i := 0; i < len(inputFiles); i++ {
		// Normalize video: scale to 1920x1080, setsar 1, fps 30, format yuv420p
		// Use force_original_aspect_ratio to keep aspect ratio and pad to fill
		vNorm := fmt.Sprintf("[%d:v]scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2,setsar=1,fps=30,format=yuv420p[v%d]", i, i)
		// Normalize audio: sample rate 44100, stereo
		aNorm := fmt.Sprintf("[%d:a]aformat=sample_rates=44100:channel_layouts=stereo[a%d]", i, i)

		filterParts = append(filterParts, vNorm, aNorm)
	}

	// Concat part
	concatFilter := ""
	for i := 0; i < len(inputFiles); i++ {
		concatFilter += fmt.Sprintf("[v%d][a%d]", i, i)
	}
	concatFilter += fmt.Sprintf("concat=n=%d:v=1:a=1[vout][aout]", len(inputFiles))

	filterParts = append(filterParts, concatFilter)
	filterComplex := strings.Join(filterParts, ";")

	args = append(args,
		"-filter_complex", filterComplex,
		"-map", "[vout]",
		"-map", "[aout]",
		"-c:v", "libx264",
		"-preset", "slow",
		"-crf", "18",
		"-c:a", "aac",
		"-b:a", "192k",
		"-y", outputPath,
	)

	return RunFFmpegCommand(args)
}
