package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	// FFmpegPath is the path to the ffmpeg executable
	FFmpegPath = "ffmpeg"
	// FFprobePath is the path to the ffprobe executable
	FFprobePath = "ffprobe"

	// RunFFmpegFunc allows overriding the ffmpeg execution logic (for testing)
	RunFFmpegFunc = runFFmpegCommandDefault
	// GetDurationFunc allows overriding the duration retrieval logic (for testing)
	GetDurationFunc = getDurationDefault
)

// RunFFmpegCommand executes an FFmpeg command
func RunFFmpegCommand(args []string) error {
	return RunFFmpegFunc(args)
}

func runFFmpegCommandDefault(args []string) error {
	cmd := exec.Command(FFmpegPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// ConcatVideoFiles concatenates multiple video files into one
func ConcatVideoFiles(videoPaths []string, outputPath string) error {
	if len(videoPaths) == 0 {
		return fmt.Errorf("no video paths provided for concatenation")
	}
	if len(videoPaths) == 1 {
		return CopyFile(videoPaths[0], outputPath)
	}

	// Create list file for FFmpeg concat demuxer
	tempDir := filepath.Dir(outputPath)
	listPath := filepath.Join(tempDir, "concat_list.txt")
	f, err := os.Create(listPath)
	if err != nil {
		return err
	}
	for _, p := range videoPaths {
		absPath, _ := filepath.Abs(p)
		f.WriteString(fmt.Sprintf("file '%s'\n", filepath.ToSlash(absPath)))
	}
	f.Close()
	defer os.Remove(listPath)

	return RunFFmpegCommand([]string{
		"-f", "concat",
		"-safe", "0",
		"-i", listPath,
		"-c", "copy",
		"-y", outputPath,
	})
}

// GetVideoDuration returns the duration of a video file in seconds
func GetVideoDuration(videoPath string) (float64, error) {
	return GetDurationFunc(videoPath)
}

// GetAudioDuration returns the duration of an audio file in seconds
func GetAudioDuration(audioPath string) (float64, error) {
	return GetDurationFunc(audioPath)
}

func getDurationDefault(path string) (float64, error) {
	cmd := exec.Command(FFprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
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

	// Handle large number of files by batching to avoid command line length limits
	// Windows has a limit of ~8191 characters, each path can be ~260. 20 files is safe.
	const batchSize = 20
	if len(inputFiles) > batchSize {
		fmt.Printf("[FFmpeg] Batching %d files into groups of %d\n", len(inputFiles), batchSize)

		var intermediateFiles []string
		dir := filepath.Dir(outputFile)

		for i := 0; i < len(inputFiles); i += batchSize {
			end := i + batchSize
			if end > len(inputFiles) {
				end = len(inputFiles)
			}

			batch := inputFiles[i:end]
			tempOutput := filepath.Join(dir, fmt.Sprintf("temp_batch_%d_%s", i, filepath.Base(outputFile)))

			// Recursively merge this batch
			if err := MergeAudioWithCrossfade(batch, tempOutput, crossfadeDuration, bitrate); err != nil {
				return fmt.Errorf("failed to merge batch %d: %w", i, err)
			}
			intermediateFiles = append(intermediateFiles, tempOutput)
		}

		// Final merge of intermediate files
		err := MergeAudioWithCrossfade(intermediateFiles, outputFile, crossfadeDuration, bitrate)

		// Cleanup intermediate files
		for _, f := range intermediateFiles {
			os.Remove(f)
		}

		if err != nil {
			return fmt.Errorf("failed to merge intermediate files: %w", err)
		}

		return nil
	}

	// Multiple files - build complex filter
	args := []string{}
	// Add input files (we already checked for empty files above, but let's be safe)
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

	// If crossfade duration is 0 or negative, use simple concat
	if crossfadeDuration <= 0 {
		filterParts := ""
		for i := 0; i < len(inputFiles); i++ {
			filterParts += fmt.Sprintf("[%d:a]", i)
		}
		filterParts += fmt.Sprintf("concat=n=%d:v=0:a=1[aout];[aout]loudnorm[final]", len(inputFiles))

		args = append(args,
			"-filter_complex", filterParts,
			"-map", "[final]",
			"-ar", "44100",
			"-ab", bitrate,
			"-y", outputFile,
		)

		return RunFFmpegCommand(args)
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
			"-preset", "medium",
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
		"-preset", "medium",
		"-crf", "18",
		"-r", strconv.Itoa(fps),
		"-y", outputFile,
	)

	return RunFFmpegCommand(args)
}

// CombineAudioVideo combines audio and video into final output
func CombineAudioVideo(videoPath, audioPath, outputPath string) error {
	// Lấy thời lượng chính xác của audio để cắt video
	audioDuration, err := GetAudioDuration(audioPath)
	if err != nil {
		return fmt.Errorf("failed to get audio duration: %w", err)
	}

	args := []string{
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy",
		"-c:a", "aac",
		"-b:a", "192k",
		"-map", "0:v:0",
		"-map", "1:a:0",
		"-t", fmt.Sprintf("%.3f", audioDuration), // Cắt chính xác theo milli-second
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
		"-preset", "medium",
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

// ConcatVideosNoAudio concatenates video-only files (no audio stream) into one MP4.
// Inputs must already be normalized to the same codec/resolution/fps.
// Used to join per-segment stock clips that were pre-rendered with -an.
func ConcatVideosNoAudio(inputFiles []string, outputPath string) error {
	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files provided")
	}

	if len(inputFiles) == 1 {
		// Single segment – just copy
		args := []string{"-i", inputFiles[0], "-c", "copy", "-y", outputPath}
		return RunFFmpegCommand(args)
	}

	// Build a concat list file
	listPath := outputPath + "_list.txt"
	f, err := os.Create(listPath)
	if err != nil {
		return fmt.Errorf("failed to create concat list: %w", err)
	}
	for _, p := range inputFiles {
		abs, err := filepath.Abs(p)
		if err != nil {
			f.Close()
			return fmt.Errorf("failed to resolve path %s: %w", p, err)
		}
		f.WriteString(fmt.Sprintf("file '%s'\n", filepath.ToSlash(abs)))
	}
	f.Close()
	defer os.Remove(listPath)

	// Use concat demuxer – fast, no re-encode when codecs match
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", listPath,
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
		"-preset", "medium",
		"-crf", "18",
		"-c:a", "aac",
		"-b:a", "192k",
		"-y", outputPath,
	)

	return RunFFmpegCommand(args)
}

// ExtractAudioSegment extracts a segment from an audio file
func ExtractAudioSegment(inputPath string, startTime float64, duration float64, outputPath string) error {
	args := []string{
		"-ss", fmt.Sprintf("%.3f", startTime),
		"-t", fmt.Sprintf("%.3f", duration),
		"-i", inputPath,
		"-c", "copy",
		"-y", outputPath,
	}
	return RunFFmpegCommand(args)
}

// RemoveAudioSilence removes silence from an audio file to improve pacing
func RemoveAudioSilence(inputPath, outputPath string) error {
	args := []string{
		"-i", inputPath,
		"-af", "silenceremove=stop_periods=-1:stop_duration=0.3:stop_threshold=-35dB",
		"-c:a", "libmp3lame",
		"-q:a", "2",
		"-y", outputPath,
	}
	return RunFFmpegCommand(args)
}

// ImageToVideo converts a static image into a video clip with Ken Burns zoom animation.
// duration: target video length in seconds. orientation: "portrait" or "landscape".
func ImageToVideo(imagePath, outputPath string, duration float64, orientation string) error {
	// Ken Burns: slow zoom from centre.
	durationSec := int(duration) + 1

	var filter string
	if orientation == "portrait" {
		// Output 1080x1920.
		// Fix jitter: Scale image up by 4x before zooming, then zoompan downcales it smoothly back to 1080x1920.
		filter = fmt.Sprintf(
			"scale=1080*4:1920*4:force_original_aspect_ratio=increase,crop=1080*4:1920*4:(iw-ow)/2:(ih-oh)/2,"+
				"zoompan=z='min(zoom+0.0007,1.15)':d=%d:x='iw/2-(iw/zoom)/2':y='ih/2-(ih/zoom)/2':s=1080x1920:fps=30,"+
				"eq=contrast=1.05:saturation=1.15:brightness=-0.02,format=yuv420p",
			durationSec*30,
		)
	} else {
		// Output 1920x1080.
		filter = fmt.Sprintf(
			"scale=1920*4:1080*4:force_original_aspect_ratio=increase,crop=1920*4:1080*4:(iw-ow)/2:(ih-oh)/2,"+
				"zoompan=z='min(zoom+0.0007,1.15)':d=%d:x='iw/2-(iw/zoom)/2':y='ih/2-(ih/zoom)/2':s=1920x1080:fps=30,"+
				"eq=contrast=1.05:saturation=1.15:brightness=-0.02,format=yuv420p",
			durationSec*30,
		)
	}

	args := []string{
		"-loop", "1",
		"-i", imagePath,
		"-vf", filter,
		"-t", fmt.Sprintf("%d", durationSec),
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "20",
		"-an",
		"-y", outputPath,
	}
	return RunFFmpegCommand(args)
}

// BurnSubtitles burns (hardcodes) subtitles from an SRT file into a video.
// orientation: "portrait" (TikTok) or "landscape" (YouTube).
func BurnSubtitles(inputPath, srtPath, outputPath, orientation string) error {
	var style string
	if orientation == "portrait" {
		// TikTok style: Vibrant yellow text, bold, slightly higher bottom position, auto-wrapping
		// MarginV=280 to be above the post interaction bar but below center
		style = "Fontname=Arial Bold,Fontsize=20,PrimaryColour=&H0000FFFF,OutlineColour=&H00000000,BorderStyle=1,Outline=2.0,Shadow=1.5,Alignment=2,MarginV=280,MarginL=60,MarginR=60,Bold=1,WrapStyle=0"
	} else {
		// YouTube style: Crisp white text, semi-bold, bottom center, auto-wrapping
		style = "Fontname=Arial Bold,Fontsize=16,PrimaryColour=&H00FFFFFF,OutlineColour=&H00000000,BorderStyle=1,Outline=1.5,Shadow=1,Alignment=2,MarginV=80,MarginL=100,MarginR=100,Bold=1,WrapStyle=0"
	}

	// FFmpeg subtitles filter needs specific escaping. Using double quotes for style to handle special chars.
	filter := fmt.Sprintf("subtitles='%s':force_style='%s'", filepath.ToSlash(srtPath), style)

	args := []string{
		"-i", inputPath,
		"-vf", filter,
		"-c:a", "copy", // keep original audio
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "20",
		"-y", outputPath,
	}

	return RunFFmpegCommand(args)
}
