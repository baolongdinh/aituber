package model

// JobCheckpoint stores the intermediate state of a video generation job.
type JobCheckpoint struct {
	JobID       string              `json:"job_id"`
	Title       string              `json:"title"`
	Segments    []CheckpointSegment `json:"segments"`
	TempDir     string              `json:"temp_dir"`
	CurrentStep string              `json:"current_step"` // "scripting", "assets", "compiling"
	Progress    int                 `json:"progress"`
	Platform    string              `json:"platform"`
	Orientation string              `json:"orientation"`
	Voice       string              `json:"voice"`
	TTSProvider string              `json:"tts_provider"`
	T2VModel    string              `json:"t2v_model"`
	T2VProvider string              `json:"t2v_provider"`
}

type CheckpointSegment struct {
	Index             int     `json:"index"`
	Text              string  `json:"text"`
	VisualPrompt      string  `json:"visual_prompt"`
	VisualDescription string  `json:"visual_description"`
	AudioPath         string  `json:"audio_path"`
	VideoPath         string  `json:"video_path"`
	AudioDone         bool    `json:"audio_done"`
	VideoDone         bool    `json:"video_done"`
	Duration          float64 `json:"duration"`
}
