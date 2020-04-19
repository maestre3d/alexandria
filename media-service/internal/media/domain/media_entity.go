package domain

import "time"

// MediaEntity Media entity type
type MediaEntity struct {
	MediaID     int64      `json:"-"`
	ExternalID  string     `json:"media_id"`
	Title       string     `json:"title"`
	DisplayName string     `json:"display_name"`
	Description *string    `json:"description"`
	UserID      string     `json:"user_id"`
	AuthorID    string     `json:"author_id"`
	PublishDate time.Time  `json:"publish_date"`
	MediaType   string     `json:"media_type"`
	TotalViews int64 `json:"total_views"`
	CreateTime  time.Time  `json:"create_time"`
	UpdateTime  time.Time  `json:"update_time"`
	DeleteTime  *time.Time `json:"-"`
	Metadata    *string    `json:"metadata,omitempty"`
	Deleted     bool       `json:"-"`
}

type MediaEntityParams struct {
	Title string
	DisplayName string
	Description *string
	UserID string
	AuthorID string
	Publi
	MediaType string
}

func NewMediaEntity()
