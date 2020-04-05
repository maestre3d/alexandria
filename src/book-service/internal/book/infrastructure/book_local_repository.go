package infrastructure

import (
	"fmt"
	"github.com/maestre3d/alexandria/src/book-service/internal/book/domain"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/util"
	"strings"
)

type BookLocalRepository struct {
	tableDB []*domain.BookEntity
	logger  util.ILogger
}

func NewBookLocalRepository(table []*domain.BookEntity, logger util.ILogger) *BookLocalRepository {
	return &BookLocalRepository{table, logger}
}

func (b *BookLocalRepository) Save(book *domain.BookEntity) error {
	book.ID = int64(len(b.tableDB)) + 1
	b.tableDB = append(b.tableDB, book)
	return nil
}

func (b *BookLocalRepository) Fetch(params *global.PaginationParams) ([]*domain.BookEntity, error) {
	if params == nil {
		params = global.NewPaginationParams("0", "0")
	} else {
		params.Sanitize()
	}

	// Index-from-limit algorithm formula
	// f(x)= w-x
	// w (omega) = x*n
	// where x = limit and n = page
	index := util.GetIndex(params.Page, params.Limit)

	if index > int64(len(b.tableDB)) {
		index = int64(len(b.tableDB))
	}

	params.Limit = params.Limit + index

	if params.Limit > int64(len(b.tableDB)) {
		params.Limit = int64(len(b.tableDB))
	}

	b.logger.Print(fmt.Sprintf("[%d:%d]", index, params.Limit), "book.infrastructure.local")

	queryResult := b.tableDB[index:params.Limit]

	return queryResult, nil
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
