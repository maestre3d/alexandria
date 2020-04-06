package application

import (
	"github.com/maestre3d/alexandria/src/book-service/internal/book/domain"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/util"
	"go.uber.org/multierr"
)

type BookUseCase struct {
	logger     util.ILogger
	repository domain.IBookRepository
}

// BookUpdateParams Required parameters to update a book
type BookUpdateParams struct {
	ID          string
	Title       string
	PublishedAt string
	UploadedBy  string
	Author      string
}

func NewBookUseCase(logger util.ILogger, repository domain.IBookRepository) *BookUseCase {
	return &BookUseCase{logger, repository}
}

func (b *BookUseCase) Create(title, publishedAt, uploadedBy, author string) error {
	bookParams := &domain.BookEntityParams{
		Title:       title,
		PublishedAt: publishedAt,
		UploadedBy:  uploadedBy,
		Author:      author,
	}
	book, err := domain.NewBookEntity(bookParams)
	if err != nil {
		return err
	}

	// Check book's title uniqueness
	existingBook, err := b.GetByTitle(book.Title)
	errors := multierr.Errors(err)
	if len(errors) > 0 {
		for _, err = range errors {
			if err != global.EntityNotFound {
				return err
			}
		}
	} else if existingBook != nil {
		return global.EntityExists
	}

	return b.repository.Save(book)
}

func (b *BookUseCase) GetByID(idString string) (*domain.BookEntity, error) {
	id, err := util.SanitizeID(idString)
	if err != nil {
		return nil, err
	}

	return b.repository.FetchByID(id)
}

func (b *BookUseCase) GetByTitle(title string) (*domain.BookEntity, error) {
	if title == "" {
		return nil, global.EmptyQuery
	}

	return b.repository.FetchByTitle(title)
}

func (b *BookUseCase) GetAll(params *global.PaginationParams) ([]*domain.BookEntity, error) {
	return b.repository.Fetch(params)
}

func (b *BookUseCase) UpdateOne(params *BookUpdateParams) error {
	id, err := util.SanitizeID(params.ID)
	if err != nil {
		return err
	}

	bookParams := &domain.BookEntityParams{
		Title:       params.Title,
		PublishedAt: params.PublishedAt,
		UploadedBy:  params.UploadedBy,
		Author:      params.Author,
	}
	book, err := domain.NewBookEntity(bookParams)
	if err != nil {
		return err
	}

	return b.repository.UpdateOne(id, book)
}

func (b *BookUseCase) RemoveOne(idString string) error {
	id, err := util.SanitizeID(idString)
	if err != nil {
		return err
	}

	return b.repository.RemoveOne(id)
}
