package infrastructure

import (
	"github.com/maestre3d/alexandria/src/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/util"
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
	media.MediaID = int64(len(m.tableDB)) + 1
	m.tableDB = append(m.tableDB, media)
	return nil
}

func (m *MediaLocalRepository) Fetch(params *global.PaginationParams) ([]*domain.MediaAggregate, error) {
	if params == nil {
		params = global.NewPaginationParams("1", "10")
	} else {
		params.Sanitize()
	}

	index := util.GetIndex(params.Page, params.Limit)

	if index > int64(len(m.tableDB)) {
		index = int64(len(m.tableDB))
	}

	params.Limit = params.Limit + index

	if params.Limit > int64(len(m.tableDB)) {
		params.Limit = int64(len(m.tableDB))
	}

	queryResult := m.tableDB[index:params.Limit]
	if len(queryResult) == 0 {
		return nil, global.EntitiesNotFound
	}

	return queryResult, nil
}

func (m *MediaLocalRepository) FetchByID(id int64) (*domain.MediaAggregate, error) {
	for _, media := range m.tableDB {
		if id == media.MediaID {
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

func (m *MediaLocalRepository) UpdateOne(id int64, mediaUpdated *domain.MediaAggregate) error {
	for _, media := range m.tableDB {
		if id == media.MediaID {
			media = mediaUpdated
			return nil
		}
	}

	return global.EntityNotFound
}

func (m *MediaLocalRepository) RemoveOne(id int64) error {
	for _, media := range m.tableDB {
		if id == media.MediaID {
			m.tableDB = m.removeIndex(m.tableDB, int(media.MediaID)-1)
			return nil
		}
	}

	return global.EntityNotFound
}

func (m *MediaLocalRepository) removeIndex(s []*domain.MediaAggregate, index int) []*domain.MediaAggregate {
	return append(s[:index], s[index+1:]...)
}
