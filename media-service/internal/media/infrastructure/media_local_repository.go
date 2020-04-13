package infrastructure

import (
	"github.com/maestre3d/alexandria/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"
	"strings"
)

type MediaLocalRepository struct {
	tableDB []*domain.MediaAggregate
	logger  util.ILogger
}

func NewMediaLocalRepository(table []*domain.MediaAggregate, logger util.ILogger) *MediaLocalRepository {
	return &MediaLocalRepository{table, logger}
}

func (m *MediaLocalRepository) Save(media *domain.MediaAggregate) error {
	for _, mediaStored := range m.tableDB {
		if mediaStored.ExternalID == media.ExternalID || mediaStored.MediaID == media.MediaID ||
			strings.ToUpper(mediaStored.Title) == strings.ToUpper(media.Title) {
			return global.EntityExists
		}
	}

	media.MediaID = int64(len(m.tableDB)) + 1
	m.tableDB = append(m.tableDB, media)
	return nil
}

func (m *MediaLocalRepository) Fetch(params *util.PaginationParams, filterMap util.FilterParams) ([]*domain.MediaAggregate, error) {
	if params == nil {
		params = util.NewPaginationParams("1", "", "10")
	} else {
		params.Sanitize()
	}

	// UPDATE: Now using keyset pagination along with page_tokens (ref. Google API Design)
	// Params.TokenID / Params.TokenUUID = page_token
	// Params.Size = page_size
	// Params.TokenID += 1 / last_item -> Params.TokenUUID = next_page_token

	/*
		OLD IMPLEMENTATION
		index := util.GetIndex(params.Page, params.Limit)

		if index > int64(len(m.tableDB)) {
			index = int64(len(m.tableDB))
		}

		params.Limit += params.Limit + index
	*/

	params.TokenID -= 1
	params.Size += 1

	if params.Size > int32(len(m.tableDB)) {
		params.Size = int32(len(m.tableDB))
	}

	if params.TokenID < 0 {
		params.TokenID = 0
	}

	queryResult := m.tableDB[int(params.TokenID):params.Size]
	if len(queryResult) == 0 {
		return nil, global.EntitiesNotFound
	}

	return queryResult, nil
}

func (m *MediaLocalRepository) FetchByID(id int64, externalID string) (*domain.MediaAggregate, error) {

	for _, media := range m.tableDB {
		// Prefer external ID instead int64 ID
		if externalID == media.ExternalID {
			return media, nil
		} else if id == media.MediaID {
			return media, nil
		}
	}

	return nil, global.EntityNotFound
}

func (m *MediaLocalRepository) FetchByTitle(title string) (*domain.MediaAggregate, error) {
	for _, media := range m.tableDB {
		if strings.ToLower(title) == strings.ToLower(media.Title) {
			return media, nil
		}
	}

	return nil, global.EntityNotFound
}

func (m *MediaLocalRepository) UpdateOne(id int64, externalID string, mediaUpdated *domain.MediaAggregate) error {
	for _, media := range m.tableDB {
		// Prefer external ID instead int64 ID
		if externalID == media.ExternalID {
			media = mediaUpdated
			return nil
		} else if id == media.MediaID {
			media = mediaUpdated
			return nil
		}
	}

	return global.EntityNotFound
}

func (m *MediaLocalRepository) RemoveOne(id int64, externalID string) error {
	for _, media := range m.tableDB {
		// Prefer external ID instead int64 ID
		if externalID == media.ExternalID {
			m.tableDB = m.removeIndex(m.tableDB, int(media.MediaID)-1)
			return nil
		} else if id == media.MediaID {
			m.tableDB = m.removeIndex(m.tableDB, int(media.MediaID)-1)
			return nil
		}
	}

	return global.EntityNotFound
}

func (m *MediaLocalRepository) removeIndex(s []*domain.MediaAggregate, index int) []*domain.MediaAggregate {
	return append(s[:index], s[index+1:]...)
}
