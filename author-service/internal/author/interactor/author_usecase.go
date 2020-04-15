package interactor

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/exception"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
	"time"
)

// AuthorUseCase Author interact actions
type AuthorUseCase struct {
	log util.ILogger
	repository domain.IAuthorRepository
}

// NewAuthorUseCase Create a new author interact
func NewAuthorUseCase(logger util.ILogger, repository domain.IAuthorRepository) *AuthorUseCase {
	return &AuthorUseCase{logger, repository}
}

// Create Store a new entity
func (u *AuthorUseCase) Create(firstName, LastName, displayName, birthDate string) (*domain.AuthorEntity, error) {
	// Validate
	birth, err := time.Parse(global.RFC3339Micro, birthDate)
	if err != nil {
		return nil, fmt.Errorf("%w:%s", exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"birth_date", global.RFC3339Micro))
	}

	author := domain.NewAuthorEntity(firstName, LastName, displayName, birth)
	err = author.IsValid()
	if err != nil {
		return nil, err
	}

	err = u.repository.Save(author)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Fire up "alexandria.author.created" domain event

	return author, nil
}

// List Get an author's list
func (u *AuthorUseCase) List(pageToken, pageSize string, filterParams util.FilterParams) ([]*domain.AuthorEntity, error) {
	return u.repository.Fetch(util.NewPaginationParams(pageToken, pageSize), filterParams)
}

// Get Obtain one author
func (u *AuthorUseCase) Get(id string) (*domain.AuthorEntity, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, exception.InvalidID
	}

	return u.repository.FetchOne(id)
}

// Update Update an author dynamically (like HTTP's PATCH)
func (u *AuthorUseCase) Update(id, firstName, LastName, displayName, birthDate string) (*domain.AuthorEntity, error) {
	// Get previous version
	author, err := u.Get(id)
	if err != nil {
		return nil, err
	}

	// Update entity dynamically
	if birthDate != "" {
		birth, err := time.Parse(global.RFC3339Micro, birthDate)
		if err != nil {
			return nil, fmt.Errorf("%w:%s", exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
				"birth_date", global.RFC3339Micro))
		}
		author.BirthDate = birth
	} else if firstName != "" {
		author.FirstName = firstName
	} else if LastName != "" {
		author.LastName = LastName
	} else if displayName != "" {
		author.DisplayName = displayName
	}

	err = author.IsValid()
	if err != nil {
		return nil, err
	}

	err = u.repository.Update(author)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Fire up "alexandria.author.updated" domain event

	return author, nil
}

// Delete Remove an author from the store
func (u *AuthorUseCase) Delete(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return exception.InvalidID
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Fire up "alexandria.author.deleted" domain event

	return u.repository.Remove(id)
}

