package model

// Series represents a multi-part video series
type Series struct {
	BaseModel
	UserID      string `gorm:"type:uuid;index;not null" json:"user_id"`
	Topic       string `gorm:"not null" json:"topic"`
	Platform    string `gorm:"not null" json:"platform"` // youtube | tiktok
	NumParts    int    `gorm:"not null" json:"num_parts"`
	Status      string `gorm:"not null;default:'processing'" json:"status"` // processing | completed | partial_failed | failed
	ContentName string `json:"content_name"`

	// Relations
	User User  `gorm:"foreignKey:UserID" json:"-"`
	Jobs []Job `gorm:"foreignKey:SeriesID" json:"jobs,omitempty"`
}

func (Series) TableName() string { return "series" }

// Job represents a single video generation task (single video or one part of a series)
type Job struct {
	BaseModel
	UserID      string  `gorm:"type:uuid;index;not null" json:"user_id"`
	SeriesID    *string `gorm:"type:uuid;index" json:"series_id,omitempty"`
	Platform    string  `gorm:"not null" json:"platform"` // youtube | tiktok
	ContentName string  `json:"content_name"`
	Topic       string  `json:"topic"`
	Voice       string  `json:"voice"`
	TTSProvider string  `json:"tts_provider"`

	// Type: "single" or "series_part"
	Type      string `gorm:"not null;default:'single'" json:"type"`
	PartIndex int    `gorm:"default:0" json:"part_index"` // 0-based, for series parts

	// Status tracking
	Status      string `gorm:"not null;default:'queued'" json:"status"` // queued | processing | completed | failed
	Progress    int    `gorm:"default:0" json:"progress"`
	CurrentStep string `json:"current_step"`

	// Output
	VideoPath string  `json:"video_path"`
	SavedPath string  `json:"saved_path"`
	ErrorMsg  *string `json:"error,omitempty"`

	// Relations
	User   User    `gorm:"foreignKey:UserID" json:"-"`
	Series *Series `gorm:"foreignKey:SeriesID" json:"-"`
}

func (Job) TableName() string { return "jobs" }
