package domain

import "github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"

// IAuthorRepository Author entity's repository
type IAuthorRepository interface {
	Save(entity *AuthorEntity) error
	Fetch(params *util.PaginationParams, filterParams util.FilterParams) ([]*AuthorEntity, error)
	FetchOne(id string) (*AuthorEntity, error)
	Update(entity *AuthorEntity) error
	Remove(id string) error
}
