package service

import (
	"aituber/config"
	"aituber/utils"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// VideoWorkflowService orchestrates the entire video creation pipeline
type VideoWorkflowService struct {
	cfg               *config.Config
	jobSvc            JobService
	textProcessor     *TextProcessor
	audioService      IAudioService
	videoProcessor    IVideoProcessor
	stockVideoService IStockVideoService
	composerService   IComposerService
	geminiService     IScriptGenerator
}

// NewVideoWorkflowService initializes workflow service with all bounded contexts
func NewVideoWorkflowService(
	cfg *config.Config,
	jobSvc JobService,
	textProcessor *TextProcessor,
	audioService IAudioService,
	videoProcessor IVideoProcessor,
	stockService IStockVideoService,
	composer IComposerService,
	gemini IScriptGenerator,
) *VideoWorkflowService {
	return &VideoWorkflowService{
		cfg:               cfg,
		jobSvc:            jobSvc,
		textProcessor:     textProcessor,
		audioService:      audioService,
		videoProcessor:    videoProcessor,
		stockVideoService: stockService,
		composerService:   composer,
		geminiService:     gemini,
	}
}

// StartGeneration kicks off background video generation pipeline
func (s *VideoWorkflowService) StartGeneration(jobID string, req GenerateRequest) {
	ctx := context.Background()
	s.jobSvc.UpdateProgress(ctx, jobID, "Creating temporary directories", 3)

	tempDir, err := utils.CreateTempDir(s.cfg.TempDir, jobID)
	if err != nil {
		s.jobSvc.MarkFailed(ctx, jobID, fmt.Errorf("failed to create temp dir: %w", err))
		return
	}

	orientation := "landscape"
	if req.Platform == "tiktok" {
		orientation = "portrait"
	}

	// 1. Script Generation
	segments, err := s.generateScript(ctx, jobID, req)
	if err != nil {
		s.jobSvc.MarkFailed(ctx, jobID, err)
		return
	}

	// 2. Parallel Segment Processing (Audio + Fetch Source Material)
	s.jobSvc.UpdateProgress(ctx, jobID, "Processing segments in parallel", 10)
	numSegments := len(segments)
	audioPaths := make([]string, numSegments)
	audioTexts := make([]string, numSegments)
	segmentVideoPaths := make([]string, numSegments)
	segmentErrors := make([]error, numSegments)

	sem := make(chan struct{}, s.cfg.MaxConcurrentTTSRequests)
	var wg sync.WaitGroup

	for i, seg := range segments {
		wg.Add(1)
		go func(index int, sSeg VideoSegment) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Sub-task: Concurrent Audio Gen and Source Material Fetch
			var aPath string
			var aErr error
			var material *StockMaterial
			var mErr error
			var wgSub sync.WaitGroup

			wgSub.Add(2)
			go func() {
				defer wgSub.Done()
				aPath, aErr = s.audioService.GenerateSingleAudio(sSeg.Text, req.Voice, -0.5, jobID, index)
			}()
			go func() {
				defer wgSub.Done()
				material, mErr = s.stockVideoService.FetchSourceMaterial(context.Background(), sSeg.VisualPrompt, sSeg.VisualDescription, req.T2VModel, req.T2VProvider, jobID, index, orientation)
			}()
			wgSub.Wait()
			// ... (rest of the loop remains same)

			if aErr != nil {
				segmentErrors[index] = fmt.Errorf("audio failed: %w", aErr)
				return
			}
			if mErr != nil {
				segmentErrors[index] = fmt.Errorf("fetch material failed: %w", mErr)
				return
			}

			audioPaths[index] = aPath
			audioTexts[index] = sSeg.Text

			// Now that we have audio, get duration and prepare video
			duration, _ := utils.GetAudioDuration(aPath)
			if duration <= 0 {
				duration = 5.0 // fallback
			}

			vPath, vErr := s.stockVideoService.PrepareVideoFromMaterial(context.Background(), material, duration, jobID, index, orientation)
			if vErr != nil {
				segmentErrors[index] = fmt.Errorf("prepare video failed: %w", vErr)
				return
			}
			segmentVideoPaths[index] = vPath
		}(i, seg)
	}

	wg.Wait()

	// Check if any critical errors occurred
	for i, err := range segmentErrors {
		if err != nil {
			s.jobSvc.MarkFailed(ctx, jobID, fmt.Errorf("segment %d failed: %w", i, err))
			return
		}
	}

	// 3. Subtitles Generation (Non-fatal)
	s.jobSvc.UpdateProgress(ctx, jobID, "Generating subtitles", 70)
	srtPath, err := s.GenerateSRT(jobID, audioPaths, audioTexts, filepath.Join(tempDir, "output"), req.Platform)
	if err != nil {
		log.Printf("[Job %s] Failed to generate subtitles: %v", jobID, err)
	}

	// 4. Merge Audio and Concatenate Videos
	s.jobSvc.UpdateProgress(ctx, jobID, "Merging assets", 80)
	mergedAudioPath := filepath.Join(tempDir, "output", "merged_audio.mp3")
	if err := s.audioService.MergeAudioFiles(audioPaths, mergedAudioPath); err != nil {
		s.jobSvc.MarkFailed(ctx, jobID, fmt.Errorf("audio merge failed: %w", err))
		return
	}

	mergedVideoPath := filepath.Join(tempDir, "output", "merged_video.mp4")
	if err := s.composerService.ConcatVideos(segmentVideoPaths, mergedVideoPath); err != nil {
		s.jobSvc.MarkFailed(ctx, jobID, fmt.Errorf("video concat failed: %w", err))
		return
	}

	// 5. Composition
	finalVideoPath, err := s.composeVideoWithAudio(ctx, jobID, tempDir, mergedVideoPath, mergedAudioPath)
	if err != nil {
		s.jobSvc.MarkFailed(ctx, jobID, err)
		return
	}

	// 6. Burn subtitles if enabled
	finalOutputPath := finalVideoPath
	if s.cfg.EnableSubtitles && srtPath != "" {
		s.jobSvc.UpdateProgress(ctx, jobID, "Burning subtitles", 90)
		subtitleVideoPath := filepath.Join(tempDir, "output", "final_video_with_subs.mp4")
		if err := utils.BurnSubtitles(finalVideoPath, srtPath, subtitleVideoPath, orientation); err == nil {
			finalOutputPath = subtitleVideoPath
		}
	}

	// 7. Save
	s.jobSvc.UpdateProgress(ctx, jobID, "Saving final output", 95)
	savedPath, _ := s.saveToOutputFolder(finalOutputPath, req.Platform, req.ContentName)

	s.jobSvc.MarkCompleted(ctx, jobID, finalOutputPath, savedPath)
	log.Printf("[Job %s] Video generation completed successfully", jobID)
}

// Sub-pipeline: Script
func (s *VideoWorkflowService) generateScript(ctx context.Context, jobID string, req GenerateRequest) ([]VideoSegment, error) {
	// 0. Use pre-provided segments if exists
	if len(req.Segments) > 0 {
		log.Printf("[Job %s] Using %d pre-provided segments", jobID, len(req.Segments))
		return req.Segments, nil
	}

	var segments []VideoSegment
	script := "" // req.Script is not in GenerateRequest, we might need to add it or handle it separately

	if script == "" {
		s.jobSvc.UpdateProgress(ctx, jobID, "Generating script with Gemini AI", 8)
		var genErr error
		if req.Platform == "tiktok" {
			segments, genErr = s.geminiService.GenerateTikTokScript(req.Topic)
		} else {
			segments, genErr = s.geminiService.GenerateYouTubeScript(req.Topic)
		}
		if genErr != nil {
			return nil, fmt.Errorf("Gemini script generation failed: %w", genErr)
		}
		log.Printf("[Job %s] Generated script (%d segments) for topic: %q", jobID, len(segments), req.Topic)
	} else {
		if len(script) > s.cfg.MaxTextLength {
			script = script[:s.cfg.MaxTextLength]
			log.Printf("[Job %s] Script truncated to %d chars", jobID, s.cfg.MaxTextLength)
		}
		chunks := s.textProcessor.SplitForSubtitles(script)
		for _, chunk := range chunks {
			segments = append(segments, VideoSegment{
				Text:         chunk,
				VisualPrompt: s.textProcessor.ExtractKeywordsFromText(chunk, req.StockKeywords),
			})
		}
		log.Printf("[Job %s] Created %d segments from direct script text", jobID, len(segments))
	}
	return segments, nil
}

func (s *VideoWorkflowService) composeVideoWithAudio(ctx context.Context, jobID, tempDir, mergedVideoPath, mergedAudioPath string) (string, error) {
	s.jobSvc.UpdateProgress(ctx, jobID, "Composing final video with audio", 90)
	composedPath := filepath.Join(tempDir, "output", "final_video_composed.mp4")
	if err := s.composerService.ComposeVideoWithAudio(mergedVideoPath, mergedAudioPath, composedPath); err != nil {
		return "", fmt.Errorf("composition failed: %w", err)
	}
	return composedPath, nil
}

func (s *VideoWorkflowService) saveToOutputFolder(srcPath, platform, contentName string) (string, error) {
	destDir := filepath.Join(s.cfg.OutputDir, platform, contentName)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output dir: %w", err)
	}
	destPath := filepath.Join(destDir, "final_video.mp4")
	if err := utils.CopyFile(srcPath, destPath); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}
	return filepath.Join("ai-videos", platform, contentName, "final_video.mp4"), nil
}

// GenerateSRT creates an SRT subtitle file based on audio durations and texts
func (s *VideoWorkflowService) GenerateSRT(jobID string, audioPaths []string, texts []string, outputDir string, platform string) (string, error) {
	srtPath := filepath.Join(outputDir, "subtitles.srt")
	file, err := os.Create(srtPath)
	if err != nil {
		return "", fmt.Errorf("failed to create SRT file: %w", err)
	}
	defer file.Close()

	currentOffset := 0.0
	if platform == "youtube" {
		if introDur, err := utils.GetVideoDuration("static/intro_video.mp4"); err == nil {
			currentOffset = introDur
		}
	}

	for i, audioPath := range audioPaths {
		if i >= len(texts) {
			break
		}
		duration, err := utils.GetAudioDuration(audioPath)
		if err != nil {
			return "", fmt.Errorf("failed to get audio duration for %s: %w", audioPath, err)
		}
		if i > 0 {
			currentOffset -= s.cfg.AudioCrossfadeDuration
		}
		start := currentOffset
		end := currentOffset + duration
		currentOffset += duration

		startStr := utils.FormatSRTTimestamp(start)
		endStr := utils.FormatSRTTimestamp(end)
		fmt.Fprintf(file, "%d\n%s --> %s\n%s\n\n", i+1, startStr, endStr, texts[i])
	}

	return srtPath, nil
}
