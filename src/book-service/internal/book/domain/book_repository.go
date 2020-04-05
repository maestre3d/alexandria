package domain

import "github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/global"

type IBookRepository interface {
	Save(book *BookEntity) error
	Fetch(params *global.PaginationParams) ([]*BookEntity, error)
	FetchByID(id int64) (*BookEntity, error)
	FetchByTitle(title string) (*BookEntity, error)
	UpdateOne(id int64, bookUpdated *BookEntity) error
	RemoveOne(id int64) error
}
