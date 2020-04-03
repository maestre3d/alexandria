package infrastructure

import (
	"github.com/maestre3d/alexandria/src/book-service/internal/book/domain"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/global"
	"strings"
)

type BookLocalRepository struct {
	tableDB []*domain.BookEntity
}

func NewBookLocalRepository(table []*domain.BookEntity) *BookLocalRepository {
	return &BookLocalRepository{table}
}

func (b *BookLocalRepository) Save(book *domain.BookEntity) error {
	book.ID = int64(len(b.tableDB)) + 1
	b.tableDB = append(b.tableDB, book)
	return nil
}

func (b *BookLocalRepository) Fetch(params *global.PaginationParams) ([]*domain.BookEntity, error) {
	return b.tableDB, nil
}

func (b *BookLocalRepository) FetchByID(id int64) (*domain.BookEntity, error) {
	for _, book := range b.tableDB {
		if id == book.ID {
			return book, nil
		}
	}

	return nil, global.EntityNotFound
}

func (b *BookLocalRepository) FetchByTitle(title string) (*domain.BookEntity, error) {
	for _, book := range b.tableDB {
		if strings.ToLower(title) == strings.ToLower(book.Title) {
			return book, nil
		}
	}

	return nil, global.EntityNotFound
}

func (b *BookLocalRepository) UpdateOne(id int64, bookUpdated *domain.BookEntity) error {
	for _, book := range b.tableDB {
		if id == book.ID {
			book = bookUpdated
			return nil
		}
	}

	return global.EntityNotFound
}

func (b *BookLocalRepository) RemoveOne(id int64) error {
	for _, book := range b.tableDB {
		if id == book.ID {
			b.tableDB = b.removeIndex(b.tableDB, int(book.ID)-1)
			return nil
		}
	}

	return global.EntityNotFound
}

func (b *BookLocalRepository) removeIndex(s []*domain.BookEntity, index int) []*domain.BookEntity {
	return append(s[:index], s[index+1:]...)
}
