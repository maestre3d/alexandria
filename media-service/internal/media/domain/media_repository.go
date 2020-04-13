package domain

import "github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"

type IMediaRepository interface {
	Save(book *MediaAggregate) error
	Fetch(params *util.PaginationParams, filterMap util.FilterParams) ([]*MediaAggregate, error)
	FetchByID(id int64, externalID string) (*MediaAggregate, error)
	FetchByTitle(title string) (*MediaAggregate, error)
	UpdateOne(id int64, externalID string, bookUpdated *MediaAggregate) error
	RemoveOne(id int64, externalID string) error
}
