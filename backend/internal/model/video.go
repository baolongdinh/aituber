package model

// Video represents a completed generated video in a user's gallery
type Video struct {
	BaseModel
	UserID      string `gorm:"type:uuid;index;not null" json:"user_id"`
	JobID       string `gorm:"type:uuid;uniqueIndex" json:"job_id"`
	Title       string `gorm:"not null" json:"title"`
	Platform    string `gorm:"not null" json:"platform"`                      // youtube | tiktok
	ContentType string `gorm:"not null;default:'single'" json:"content_type"` // single | series_part

	// File info
	FilePath     string `json:"file_path"`
	ThumbnailURL string `json:"thumbnail_url"`
	DurationSec  int    `json:"duration_sec"`

	// Explore / public share
	IsPublic  bool  `gorm:"default:false" json:"is_public"`
	ViewCount int64 `gorm:"default:0" json:"view_count"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"-"`
	Job  Job  `gorm:"foreignKey:JobID" json:"-"`
}

func (Video) TableName() string { return "videos" }
