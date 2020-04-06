package domain

import (
	"errors"
	"fmt"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/global"
	"go.uber.org/multierr"
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

	// Validate params
	var publishedAt time.Time
	var err error

	if params.PublishedAt != "" {
		publishedAt, err = time.Parse(global.RFC3339Micro, params.PublishedAt)
		if err != nil {
			return nil, multierr.Append(global.EntityDomainError, fmt.Errorf("%s: %s", global.InvalidFieldFormat.Error(), "published_at"))
		}
	} else {
		return nil, multierr.Append(global.EntityDomainError, errors.New("published_at is required"))
	}

	var uploadedBy int64
	if params.UploadedBy != "" {
		uploadedBy, err = strconv.ParseInt(params.UploadedBy, 10, 64)
		if err != nil {
			return nil, multierr.Append(global.EntityDomainError, fmt.Errorf("%s: %s", global.InvalidFieldFormat.Error(), "uploaded_by"))
		}
	} else {
		return nil, multierr.Append(global.EntityDomainError, errors.New("uploaded_by is required"))
	}

	var author int64
	if params.Author != "" {
		author, err = strconv.ParseInt(params.Author, 10, 64)
		if err != nil {
			return nil, multierr.Append(global.EntityDomainError, fmt.Errorf("%s: %s", global.InvalidFieldFormat.Error(), "author"))
		}
	} else {
		return nil, multierr.Append(global.EntityDomainError, errors.New("author is required"))
	}

	// Cleaned params, proceed with book mapping
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
		return multierr.Append(global.EntityDomainError, errors.New("invalid user ID"))
	} else if b.Author <= 0 {
		return multierr.Append(global.EntityDomainError, errors.New("invalid author ID"))
	} else if b.Title == "" || len(b.Title) >= 100 {
		// Validate title
		return multierr.Append(global.EntityDomainError, errors.New("invalid title"))
	}

	return nil
}
