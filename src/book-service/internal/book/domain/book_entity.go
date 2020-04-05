package domain

import (
	"errors"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/global"
	"strconv"
	"time"
)

// BookEntity Book entity model
type BookEntity struct {
	ID          int64     `json:"id"`
	Title       string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UploadedBy  int64     `json:"created_by"`
	Author      int64     `json:"author"`
}

// BookEntityParams Required parameters to create an entity
type BookEntityParams struct {
	Title       string
	PublishedAt string
	UploadedBy  string
	Author      string
}

func NewBookEntity(params *BookEntityParams) (*BookEntity, error) {
	var publishedAt time.Time
	var err error

	if params.PublishedAt != "" {
		publishedAt, err = time.Parse(global.RFC3339Micro, params.PublishedAt)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("published_at is required")
	}

	uploadedBy, err := strconv.ParseInt(params.UploadedBy, 10, 64)
	if err != nil {
		return nil, err
	}

	author, err := strconv.ParseInt(params.Author, 10, 64)
	if err != nil {
		return nil, err
	}

	book := &BookEntity{
		ID:          0,
		Title:       params.Title,
		PublishedAt: publishedAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UploadedBy:  uploadedBy,
		Author:      author,
	}

	err = book.Validate()
	if err != nil {
		return nil, err
	}

	return book, nil
}

func (b *BookEntity) Validate() error {
	// Validate FK are not empty
	if b.UploadedBy <= 0 {
		return errors.New("invalid user ID")
	} else if b.Author <= 0 {
		return errors.New("invalid author ID")
	} else if b.Title == "" || len(b.Title) >= 100 {
		// Validate name
		return errors.New("invalid name")
	}

	return nil
}
