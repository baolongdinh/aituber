package service

import (
	"aituber/config"
	"aituber/internal/model"
	"aituber/utils"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type VideoWorkflowService struct {
	cfg               *config.Config
	jobSvc            JobService
	textProcessor     *TextProcessor
	audioService      IAudioService
	videoProcessor    IVideoProcessor
	stockVideoService IStockVideoService
	composerService   IComposerService
	geminiService     IScriptGenerator
	activeJobs        sync.Map // map[string]context.CancelFunc
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.activeJobs.Store(jobID, cancel)
	defer s.activeJobs.Delete(jobID)

	// 0. Load or Initialize Checkpoint
	checkpoint, err := s.jobSvc.GetCheckpoint(ctx, jobID)
	if err != nil {
		log.Printf("[Job %s] Warning: failed to load checkpoint: %v", jobID, err)
	}

	if checkpoint == nil {
		checkpoint = &model.JobCheckpoint{
			JobID:       jobID,
			Platform:    req.Platform,
			Voice:       req.Voice,
			TTSProvider: req.TTSProvider,
			T2VModel:    req.T2VModel,
			T2VProvider: req.T2VProvider,
		}
	}

	s.jobSvc.UpdateProgress(ctx, jobID, "Creating temporary directories", 3)

	tempDir := checkpoint.TempDir
	if tempDir == "" {
		tempDir, err = utils.CreateTempDir(s.cfg.TempDir, jobID)
		if err != nil {
			s.jobSvc.MarkFailed(ctx, jobID, fmt.Errorf("failed to create temp dir: %w", err))
			return
		}
		checkpoint.TempDir = tempDir
		s.jobSvc.SaveCheckpoint(ctx, jobID, checkpoint)
	}

	orientation := "landscape"
	if checkpoint.Platform == "tiktok" {
		orientation = "portrait"
	}
	checkpoint.Orientation = orientation

	// 1. Script Generation
	if len(checkpoint.Segments) == 0 {
		segments, aiTitle, err := s.generateScript(ctx, jobID, req)
		if err != nil {
			s.jobSvc.MarkFailed(ctx, jobID, err)
			return
		}

		// Update job title if AI generated a better one
		if aiTitle != "" {
			s.jobSvc.UpdateJobTitle(ctx, jobID, aiTitle)
			checkpoint.Title = aiTitle
		} else if req.ContentName != "" {
			checkpoint.Title = req.ContentName
		}

		// Initialize segments in checkpoint
		for i, seg := range segments {
			checkpoint.Segments = append(checkpoint.Segments, model.CheckpointSegment{
				Index:             i,
				Text:              seg.Text,
				VisualPrompt:      seg.VisualPrompt,
				VisualDescription: seg.VisualDescription,
			})
		}
		s.jobSvc.SaveCheckpoint(ctx, jobID, checkpoint)
	}

	// 2. Parallel Segment Processing (Audio + Fetch Source Material)
	s.jobSvc.UpdateProgress(ctx, jobID, "Processing segments in parallel", 10)
	numSegments := len(checkpoint.Segments)
	audioPaths := make([]string, numSegments)
	audioTexts := make([]string, numSegments)
	segmentVideoPaths := make([]string, numSegments)
	segmentErrors := make([]error, numSegments)

	sem := make(chan struct{}, s.cfg.MaxConcurrentTTSRequests)
	var wg sync.WaitGroup
	var cpMu sync.Mutex

	for i := range checkpoint.Segments {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			cpMu.Lock()
			seg := &checkpoint.Segments[index]
			if seg.AudioDone && seg.VideoDone {
				audioPaths[index] = seg.AudioPath
				audioTexts[index] = seg.Text
				segmentVideoPaths[index] = seg.VideoPath
				cpMu.Unlock()
				return
			}
			checkpointVoice := checkpoint.Voice
			checkpointOrientation := checkpoint.Orientation
			checkpointT2VModel := checkpoint.T2VModel
			checkpointT2VProvider := checkpoint.T2VProvider
			cpMu.Unlock()

			sem <- struct{}{}
			defer func() { <-sem }()

			// Sub-task: Concurrent Audio Gen and Source Material Fetch
			var aPath string
			var aErr error
			var material *StockMaterial
			var mErr error
			var wgSub sync.WaitGroup

			audioNeeded := false
			videoNeeded := false

			cpMu.Lock()
			if !seg.AudioDone {
				audioNeeded = true
			}
			if !seg.VideoDone {
				videoNeeded = true
			}
			cpMu.Unlock()

			if audioNeeded {
				wgSub.Add(1)
				go func() {
					defer wgSub.Done()
					aPath, aErr = s.audioService.GenerateSingleAudio(seg.Text, checkpointVoice, checkpoint.TTSProvider, -0.5, jobID, index)
				}()
			}

			if videoNeeded {
				wgSub.Add(1)
				go func() {
					defer wgSub.Done()
					material, mErr = s.stockVideoService.FetchSourceMaterial(ctx, seg.VisualPrompt, seg.VisualDescription, checkpointT2VModel, checkpointT2VProvider, jobID, index, checkpointOrientation)
				}()
			}
			wgSub.Wait()

			if aErr != nil {
				segmentErrors[index] = fmt.Errorf("audio failed: %w", aErr)
				return
			}
			if mErr != nil {
				segmentErrors[index] = fmt.Errorf("fetch material failed: %w", mErr)
				return
			}

			cpMu.Lock()
			if audioNeeded {
				seg.AudioPath = aPath
				seg.AudioDone = true
				dur, _ := utils.GetAudioDuration(aPath)
				seg.Duration = dur
			}
			duration := seg.Duration
			cpMu.Unlock()

			audioPaths[index] = seg.AudioPath
			audioTexts[index] = seg.Text

			if videoNeeded {
				vPath, vErr := s.stockVideoService.PrepareVideoFromMaterial(ctx, material, duration, jobID, index, checkpointOrientation)
				if vErr != nil {
					segmentErrors[index] = fmt.Errorf("prepare video failed: %w", vErr)
					return
				}
				cpMu.Lock()
				seg.VideoPath = vPath
				seg.VideoDone = true
				cpMu.Unlock()
			}
			segmentVideoPaths[index] = seg.VideoPath

			// Periodic save (every segment finished)
			s.jobSvc.SaveCheckpoint(ctx, jobID, checkpoint)
		}(i)
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
	srtPath, err := s.GenerateSRT(jobID, audioPaths, audioTexts, filepath.Join(tempDir, "output"), checkpoint.Platform)
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
		if err := utils.BurnSubtitles(finalVideoPath, srtPath, subtitleVideoPath, checkpoint.Orientation); err == nil {
			finalOutputPath = subtitleVideoPath
		}
	}

	// 7. Save & Thumbnail Extraction
	s.jobSvc.UpdateProgress(ctx, jobID, "Saving final output & extracting thumbnail", 95)
	savedPath, _ := s.saveToOutputFolder(finalOutputPath, checkpoint.Platform, checkpoint.Title)

	// Extract and save thumbnail
	tempThumbPath := filepath.Join(tempDir, "output", "thumbnail.jpg")
	var savedThumbPath string
	if err := s.composerService.ExtractThumbnail(finalOutputPath, tempThumbPath, 1.0); err == nil {
		savedThumbPath, _ = s.saveThumbnailToOutputFolder(tempThumbPath, checkpoint.Platform, checkpoint.Title)
	}

	s.jobSvc.MarkCompleted(ctx, jobID, finalOutputPath, savedPath, savedThumbPath)
	log.Printf("[Job %s] Video generation completed successfully", jobID)
}

// CancelJob terminates an active job
func (s *VideoWorkflowService) CancelJob(jobID string) bool {
	if cancel, ok := s.activeJobs.Load(jobID); ok {
		cancel.(context.CancelFunc)()
		s.activeJobs.Delete(jobID)
		log.Printf("[Job %s] Job cancellation requested", jobID)
		return true
	}
	return false
}

// Sub-pipeline: Script
func (s *VideoWorkflowService) generateScript(ctx context.Context, jobID string, req GenerateRequest) ([]VideoSegment, string, error) {
	// 0. Use pre-provided segments if exists
	if len(req.Segments) > 0 {
		log.Printf("[Job %s] Using %d pre-provided segments", jobID, len(req.Segments))
		return req.Segments, "", nil
	}

	var segments []VideoSegment
	var title string
	script := ""

	if script == "" {
		s.jobSvc.UpdateProgress(ctx, jobID, "Generating script with Gemini AI", 8)
		var genScript *GeneratedScript
		var genErr error
		if req.Platform == "tiktok" {
			genScript, genErr = s.geminiService.GenerateTikTokScript(req.Topic)
		} else {
			genScript, genErr = s.geminiService.GenerateYouTubeScript(req.Topic)
		}
		if genErr != nil {
			return nil, "", fmt.Errorf("Gemini script generation failed: %w", genErr)
		}
		segments = genScript.Segments
		title = genScript.Title
		log.Printf("[Job %s] Generated script (%d segments) with title %q for topic: %q", jobID, len(segments), title, req.Topic)
	} else {
		// ... (legacy handling for manual script, returns empty title)
		if len(script) > s.cfg.MaxTextLength {
			script = script[:s.cfg.MaxTextLength]
		}
		chunks := s.textProcessor.SplitForSubtitles(script)
		for _, chunk := range chunks {
			segments = append(segments, VideoSegment{
				Text:         chunk,
				VisualPrompt: s.textProcessor.ExtractKeywordsFromText(chunk, req.StockKeywords),
			})
		}
	}
	return segments, title, nil
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
	return filepath.Join("/", "ai-videos", platform, contentName, "final_video.mp4"), nil
}

func (s *VideoWorkflowService) saveThumbnailToOutputFolder(srcPath, platform, contentName string) (string, error) {
	destDir := filepath.Join(s.cfg.OutputDir, platform, contentName)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output dir: %w", err)
	}
	destPath := filepath.Join(destDir, "thumbnail.jpg")
	if err := utils.CopyFile(srcPath, destPath); err != nil {
		return "", fmt.Errorf("failed to copy thumbnail: %w", err)
	}
	return filepath.Join("/", "ai-videos", platform, contentName, "thumbnail.jpg"), nil
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
