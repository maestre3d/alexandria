package domain

import "github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"

type IMediaRepository interface {
	Save(entity *MediaEntity) error
	Fetch(params *util.PaginationParams, filterMap util.FilterParams) ([]*MediaEntity, error)
	FetchByID(id string) (*MediaEntity, error)
	Update(entity *MediaEntity) error
	Remove(id string) error
}
