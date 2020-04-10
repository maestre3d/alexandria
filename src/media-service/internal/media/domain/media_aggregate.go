package domain

import "time"

// MediaAggregate Media aggregate type
type MediaAggregate struct {
	MediaID     int64      `json:"-"`
	ExternalID  string     `json:"media_id"`
	Title       string     `json:"title"`
	DisplayName string     `json:"display_name"`
	Description *string    `json:"description"`
	UserID      string     `json:"user_id"`
	AuthorID    string     `json:"author_id"`
	PublishDate time.Time  `json:"publish_date"`
	MediaType   string     `json:"media_type"`
	CreateTime  time.Time  `json:"create_time"`
	UpdateTime  time.Time  `json:"update_time"`
	DeleteTime  *time.Time `json:"-"`
	Metadata    *string    `json:"-"`
	Deleted     bool       `json:"-"`
}

func (m *MediaAggregate) ToMediaEntity() *MediaEntity {
	return &MediaEntity{
		MediaID:     mediaID{m.MediaID},
		ExternalID:  externalID{m.ExternalID},
		Title:       title{m.Title},
		DisplayName: displayName{m.DisplayName},
		Description: &description{m.Description},
		UserID:      userID{m.UserID},
		AuthorID:    authorID{m.AuthorID},
		PublishDate: m.PublishDate,
		MediaType:   mediaType{m.MediaType},
		CreateTime:  m.CreateTime,
		UpdateTime:  m.UpdateTime,
		DeleteTime:  m.DeleteTime,
		Metadata:    m.Metadata,
		Deleted:     m.Deleted,
	}
}
