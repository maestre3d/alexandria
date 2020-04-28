package interactor

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/exception"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
)

// AuthorUseCase Author interact actions
type AuthorUseCase struct {
	log        log.Logger
	repository domain.IAuthorRepository
	eventBus domain.IAuthorEventBus
}

// NewAuthorUseCase Create a new author interact
func NewAuthorUseCase(logger log.Logger, repository domain.IAuthorRepository, bus domain.IAuthorEventBus) *AuthorUseCase {
	return &AuthorUseCase{logger, repository, bus}
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

	// Ensure display_name uniqueness, as a username
	existingAuthors, _, err := u.List("0", "1", util.FilterParams{"display_name": displayName})
	if err == nil && len(existingAuthors) > 0 {
		return nil, exception.EntityExists
	}

	err = u.repository.Save(author)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Fire up "alexandria.author.created" domain event
	go func() {
		err = u.eventBus.AuthorCreated(author)
		if err != nil {
			u.log.Log("method", "author.create", "err", err.Error())
		} else {
			u.log.Log("method", "author.create", "msg", "event alexandria.author.created published")
		}
	}()

	return author, nil
}

// List Get an author's list
func (u *AuthorUseCase) List(pageToken, pageSize string, filterParams util.FilterParams) (output []*domain.AuthorEntity, nextToken string, err error) {
	params := util.NewPaginationParams(pageToken, pageSize)
	output, err = u.repository.Fetch(params, filterParams)

	nextToken = ""
	if len(output) >= params.Size {
		nextToken = output[len(output)-1].ExternalID
		output = output[0 : len(output)-1]
	}
	return
}

// Get Obtain one author
func (u *AuthorUseCase) Get(id string) (*domain.AuthorEntity, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, exception.InvalidID
	}

	return u.repository.FetchByID(id)
}

// Update Update an author dynamically (like HTTP's PATCH)
func (u *AuthorUseCase) Update(id, firstName, lastName, displayName, birthDate string) (*domain.AuthorEntity, error) {
	// Check if body has values
	if firstName == "" && lastName == "" && displayName == "" && birthDate == "" {
		return nil, exception.EmptyBody
	}

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
	} else if lastName != "" {
		author.LastName = lastName
	} else if displayName != "" {
		existingAuthors, _, err := u.List("0", "1", util.FilterParams{"display_name": displayName})
		if err == nil && len(existingAuthors) > 0 {
			return nil, exception.EntityExists
		}

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

	err = u.repository.Remove(id)
	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Fire up "alexandria.author.deleted" domain event
	go func() {
		if err == nil {
			if errBus := u.eventBus.AuthorDeleted(id); errBus != nil {
				u.log.Log("method", "author.delete", "err", errBus.Error())
			} else {
				u.log.Log("method", "author.delete", "msg", "event alexandria.author.deleted published")
			}
		}
	}()

	return err
}
