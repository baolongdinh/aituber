package services

import (
	"aituber/config"
	"aituber/models"
	"aituber/utils"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// VideoWorkflowService orchestrates the entire video creation pipeline
type VideoWorkflowService struct {
	cfg               *config.Config
	jobManager        IJobManager
	textProcessor     *TextProcessor
	audioService      IAudioService
	videoService      *VideoService // We keep concrete for now if not heavily mocked
	stockVideoService IStockVideoService
	composerService   IComposerService
	geminiService     IScriptGenerator
}

// NewVideoWorkflowService initializes workflow service with all bounded contexts
func NewVideoWorkflowService(
	cfg *config.Config,
	jobManager IJobManager,
	textProcessor *TextProcessor,
	audioService IAudioService,
	videoService *VideoService,
	stockService IStockVideoService,
	composer IComposerService,
	gemini IScriptGenerator,
) *VideoWorkflowService {
	return &VideoWorkflowService{
		cfg:               cfg,
		jobManager:        jobManager,
		textProcessor:     textProcessor,
		audioService:      audioService,
		videoService:      videoService,
		stockVideoService: stockService,
		composerService:   composer,
		geminiService:     gemini,
	}
}

// StartGeneration kicks off background video generation pipeline
func (s *VideoWorkflowService) StartGeneration(jobID string, req models.GenerateRequest) {
	s.jobManager.UpdateProgress(jobID, "Creating temporary directories", 3)

	tempDir, err := utils.CreateTempDir(s.cfg.TempDir, jobID)
	if err != nil {
		s.jobManager.MarkFailed(jobID, fmt.Errorf("failed to create temp dir: %w", err))
		return
	}

	orientation := "landscape"
	if req.Platform == "tiktok" {
		orientation = "portrait"
	}

	// 1. Script Generation
	segments, err := s.generateScript(jobID, req)
	if err != nil {
		s.jobManager.MarkFailed(jobID, err)
		return
	}

	// 2. Audio Generation
	audioPaths, audioTexts, err := s.generateAudio(jobID, req, segments)
	if err != nil {
		s.jobManager.MarkFailed(jobID, err)
		return
	}

	// 3. Subtitles Generation (Non-fatal)
	s.jobManager.UpdateProgress(jobID, "Generating subtitles", 32)
	if _, err := s.generateSRT(jobID, audioPaths, audioTexts, filepath.Join(tempDir, "output"), req.Platform); err != nil {
		log.Printf("[Job %s] Failed to generate subtitles: %v", jobID, err)
	}

	// 4. Merge Audio
	mergedAudioPath, err := s.mergeAudio(jobID, tempDir, audioPaths)
	if err != nil {
		s.jobManager.MarkFailed(jobID, err)
		return
	}

	// 5. Stock Video Gathering
	mergedVideoPath, err := s.gatherAndConcatStockVideos(jobID, tempDir, segments, audioPaths, req, orientation)
	if err != nil {
		s.jobManager.MarkFailed(jobID, err)
		return
	}

	// 6. Composition
	finalVideoPath, err := s.composeVideoWithAudio(jobID, tempDir, mergedVideoPath, mergedAudioPath)
	if err != nil {
		s.jobManager.MarkFailed(jobID, err)
		return
	}

	// 7. Add Intro/Outro for YouTube
	finalVideoPath, err = s.addIntroOutro(jobID, tempDir, finalVideoPath, req.Platform)
	if err != nil {
		s.jobManager.MarkFailed(jobID, err)
		return
	}

	// 8. Save
	s.jobManager.UpdateProgress(jobID, "Saving video to output folder", 98)
	savedPath, err := s.saveToOutputFolder(finalVideoPath, req.Platform, req.ContentName)
	if err != nil {
		log.Printf("[Job %s] Warning: could not save to output folder: %v", jobID, err)
		savedPath = ""
	} else {
		log.Printf("[Job %s] Video saved to: %s", jobID, savedPath)
	}

	s.jobManager.UpdateProgress(jobID, "Complete", 100)
	s.jobManager.MarkCompleted(jobID, finalVideoPath, savedPath)
	log.Printf("[Job %s] Video generation completed successfully", jobID)
}

// Sub-pipeline: Script
func (s *VideoWorkflowService) generateScript(jobID string, req models.GenerateRequest) ([]models.VideoSegment, error) {
	var segments []models.VideoSegment
	script := req.Script

	if script == "" {
		s.jobManager.UpdateProgress(jobID, "Generating script with Gemini AI", 8)
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
			segments = append(segments, models.VideoSegment{
				Text:         chunk,
				VisualPrompt: s.textProcessor.ExtractKeywordsFromText(chunk, req.StockKeywords),
			})
		}
		log.Printf("[Job %s] Created %d segments from direct script text", jobID, len(segments))
	}
	return segments, nil
}

// Sub-pipeline: Audio
func (s *VideoWorkflowService) generateAudio(jobID string, req models.GenerateRequest, segments []models.VideoSegment) ([]string, []string, error) {
	s.jobManager.UpdateProgress(jobID, "Preparing text for audio generation", 12)
	var audioTexts []string
	for _, seg := range segments {
		if strings.TrimSpace(seg.Text) != "" {
			audioTexts = append(audioTexts, seg.Text)
		}
	}

	if len(audioTexts) == 0 {
		return nil, nil, fmt.Errorf("no valid script segments extracted to process")
	}

	s.jobManager.UpdateProgress(jobID, fmt.Sprintf("Generating %d audio chunks", len(audioTexts)), 20)
	audioPaths, err := s.audioService.GenerateAudioChunks(
		audioTexts,
		req.Voice,
		req.SpeakingSpeed,
		jobID,
		s.cfg.MaxConcurrentTTSRequests,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("audio generation failed: %w", err)
	}
	return audioPaths, audioTexts, nil
}

// Sub-pipeline: Merge Audio
func (s *VideoWorkflowService) mergeAudio(jobID, tempDir string, audioPaths []string) (string, error) {
	s.jobManager.UpdateProgress(jobID, "Merging audio", 42)
	mergedAudioPath := filepath.Join(tempDir, "output", "merged_audio.mp3")
	if err := s.audioService.MergeAudioFiles(audioPaths, mergedAudioPath); err != nil {
		return "", fmt.Errorf("audio merge failed: %w", err)
	}
	return mergedAudioPath, nil
}

// Sub-pipeline: Stock Video
func (s *VideoWorkflowService) gatherAndConcatStockVideos(
	jobID, tempDir string, segments []models.VideoSegment, audioPaths []string,
	req models.GenerateRequest, orientation string,
) (string, error) {
	s.jobManager.UpdateProgress(jobID, "Preparing per-segment stock videos", 50)

	realDurations := make([]float64, len(audioPaths))
	for i, ap := range audioPaths {
		d, err := utils.GetAudioDuration(ap)
		if err != nil {
			log.Printf("[Job %s] Could not get duration of chunk %d: %v (using estimate 5s)", jobID, i, err)
			d = 5.0
		}
		realDurations[i] = d
	}

	segKeywords := make([]string, len(segments))
	for i, seg := range segments {
		segKeywords[i] = seg.VisualPrompt
		if strings.TrimSpace(segKeywords[i]) == "" {
			segKeywords[i] = s.textProcessor.ExtractKeywordsFromText(seg.Text, req.StockKeywords)
		}
	}

	segVideoPaths := make([]string, len(segments))
	segErrors := make([]error, len(segments))
	sem := make(chan struct{}, 3)
	var wg sync.WaitGroup

	for i := range segments {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			s.jobManager.UpdateProgress(jobID, fmt.Sprintf("Fetching stock video for segment %d/%d", idx+1, len(segments)), 50+idx*30/len(segments))

			vp, err := s.stockVideoService.PrepareSegmentVideo(
				segKeywords[idx],
				realDurations[idx],
				jobID,
				idx,
				orientation,
			)
			if err != nil {
				segErrors[idx] = err
				log.Printf("[Job %s] Segment %d video error: %v", jobID, idx, err)
			} else {
				segVideoPaths[idx] = vp
			}
		}(i)
	}
	wg.Wait()

	var goodSegPaths []string
	for i, err := range segErrors {
		if err != nil {
			log.Printf("[Job %s] Segment %d failed, skipping from timeline: %v", jobID, i, err)
			continue
		}
		if segVideoPaths[i] != "" {
			goodSegPaths = append(goodSegPaths, segVideoPaths[i])
		}
	}

	if len(goodSegPaths) == 0 {
		return "", fmt.Errorf("all segment video fetches failed")
	}

	s.jobManager.UpdateProgress(jobID, "Concatenating segment videos", 82)
	concatVideoPath := filepath.Join(tempDir, "output", "segments_concat.mp4")
	if err := utils.ConcatVideosNoAudio(goodSegPaths, concatVideoPath); err != nil {
		return "", fmt.Errorf("segment video concat failed: %w", err)
	}

	return concatVideoPath, nil
}

// Sub-pipeline: Compositing
func (s *VideoWorkflowService) composeVideoWithAudio(jobID, tempDir, mergedVideoPath, mergedAudioPath string) (string, error) {
	s.jobManager.UpdateProgress(jobID, "Composing final video with audio", 90)
	composedPath := filepath.Join(tempDir, "output", "final_video_composed.mp4")
	if err := s.composerService.ComposeVideoWithAudio(mergedVideoPath, mergedAudioPath, composedPath); err != nil {
		return "", fmt.Errorf("composition failed: %w", err)
	}
	return composedPath, nil
}

// Sub-pipeline: Intro Outro
func (s *VideoWorkflowService) addIntroOutro(jobID, tempDir, finalVideoPath, platform string) (string, error) {
	s.jobManager.UpdateProgress(jobID, "Adding intro/outro", 95)

	introPath := "static/intro_video.mp4"
	outroPath := "static/outro_video.mp4"

	concatList := utils.BuildFinalConcatList(platform, introPath, outroPath, finalVideoPath)

	if len(concatList) > 1 {
		finalWithIntroOutro := filepath.Join(tempDir, "output", "final_complete.mp4")
		if err := utils.ConcatVideos(concatList, finalWithIntroOutro); err != nil {
			return "", fmt.Errorf("failed to add intro/outro: %w", err)
		}
		return finalWithIntroOutro, nil
	}

	return finalVideoPath, nil
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

func (s *VideoWorkflowService) generateSRT(jobID string, audioPaths []string, texts []string, outputDir string, platform string) (string, error) {
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
