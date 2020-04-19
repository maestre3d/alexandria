package service

import (
	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
)

type IAuthorService interface {
	Create(firstName, lastName, displayName, birthDate string) (*domain.AuthorEntity, error)
	List(pageToken, pageSize string, filterParams util.FilterParams) ([]*domain.AuthorEntity, string, error)
	Get(id string) (*domain.AuthorEntity, error)
	Update(id, firstName, lastName, displayName, birthDate string) (*domain.AuthorEntity, error)
	Delete(id string) error
}
