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

// AuthorService Author interact actions
type AuthorUseCase struct {
	log util.ILogger
	repository domain.IAuthorRepository
}

// NewAuthorUseCase Create a new author interact
func NewAuthorUseCase(logger util.ILogger, repository domain.IAuthorRepository) *AuthorUseCase {
	return &AuthorUseCase{logger, repository}
}

// Create Store a new entity
func (s *AuthorUseCase) Create(firstName, LastName, displayName, birthDate string) (*domain.AuthorEntity, error) {
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

	// Ensure display_name uniqueness, as a username
	existingAuthors, _, err := s.List("0", "1", util.FilterParams{"display_name":displayName})
	if err == nil && len(existingAuthors) > 0 {
		return nil, exception.EntityExists
	}

	err = s.repository.Save(author)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Fire up "alexandria.author.created" domain event

	return author, nil
}

// List Get an author's list
func (s *AuthorUseCase) List(pageToken, pageSize string, filterParams util.FilterParams) (output []*domain.AuthorEntity, nextToken string, err error) {
	params := util.NewPaginationParams(pageToken, pageSize)
	output, err = s.repository.Fetch(params, filterParams)

	nextToken = ""
	if len(output) >= params.Size {
		nextToken = output[len(output)-1].ExternalID
		output = output[0:len(output)-1]
	}
	return
}

// Get Obtain one author
func (s *AuthorUseCase) Get(id string) (*domain.AuthorEntity, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, exception.InvalidID
	}

	return s.repository.FetchOne(id)
}

// Update Update an author dynamically (like HTTP's PATCH)
func (s *AuthorUseCase) Update(id, firstName, lastName, displayName, birthDate string) (*domain.AuthorEntity, error) {
	// Get previous version
	author, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	// Check if body has values
	if firstName == "" && lastName == "" && displayName == "" && birthDate == "" {
		return nil, exception.EmptyBody
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
	} else if lastName != "" {
		author.LastName = lastName
	} else if displayName != "" {
		existingAuthors, _, err := s.List("0", "1", util.FilterParams{"display_name":displayName})
		if err == nil && len(existingAuthors) > 0 {
			return nil, exception.EntityExists
		}

		author.DisplayName = displayName
	}

	err = author.IsValid()
	if err != nil {
		return nil, err
	}

	err = s.repository.Update(author)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Fire up "alexandria.author.updated" domain event

	return author, nil
}

// Delete Remove an author from the store
func (s *AuthorUseCase) Delete(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return exception.InvalidID
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Fire up "alexandria.author.deleted" domain event

	return s.repository.Remove(id)
}

