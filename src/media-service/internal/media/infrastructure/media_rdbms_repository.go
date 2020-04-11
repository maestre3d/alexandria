package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"strings"
	"time"

	"github.com/maestre3d/alexandria/src/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/util"
)

type MediaRDBMSRepository struct {
	db     *sql.DB
	logger util.ILogger
	ctx    context.Context
}

func NewMediaRDBMSRepository(db *sql.DB, logger util.ILogger, ctx context.Context) *MediaRDBMSRepository {
	return &MediaRDBMSRepository{
		db,
		logger,
		ctx,
	}
}

func (m *MediaRDBMSRepository) Save(media *domain.MediaAggregate) error {
	conn, err := m.db.Conn(m.ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()

	statement := `INSERT INTO MEDIA(EXTERNAL_ID, TITLE, DISPLAY_NAME, DESCRIPTION, USER_ID, AUTHOR_ID, PUBLISH_DATE, MEDIA_TYPE) VALUES
	($1, $2, $3, $4, $5, $6, $7, $8);`

	_, err = conn.ExecContext(m.ctx, statement, media.ExternalID, media.Title, media.DisplayName, media.Description, media.UserID, media.AuthorID,
		media.PublishDate, media.MediaType)

	if customErr, ok := err.(*pq.Error); ok {
		if customErr.Code == "23505" {
			return global.EntityExists
		}
	}

	return err
}

func (m *MediaRDBMSRepository) Fetch(params *util.PaginationParams, filterMap util.FilterParams) ([]*domain.MediaAggregate, error) {
	conn, err := m.db.Conn(m.ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = conn.Close()
	}()

	if params == nil {
		params = util.NewPaginationParams("1", "", "10")
	}

	// UPDATE: Now using keyset pagination along with page_tokens (ref. Google API Design)
	// Params.TokenID / Params.TokenUUID = page_token
	// Params.Size = page_size
	// Params.TokenID += 1 / last_item -> Params.TokenUUID = next_page_token

	// index := util.GetIndex(params.Page, params.Limit)

	// Dynamic query builder using Criteria pattern
	// Add header
	statement := `SELECT * FROM MEDIA WHERE `
	params.Size += 1

	// Criteria filtering
	for filterKey, value := range filterMap {
		switch {
		case filterKey == "query" && value != "":
			statement += AndCriteria(QueryCriteria(value))
		case filterKey == "author" && value != "":
			statement += AndCriteria(AuthorCriteria(value))
		case filterKey == "user" && value != "":
			statement += AndCriteria(PublisherCriteria(value))
		case filterKey == "media" && value != "":
			statement += AndCriteria(MediaTypeCriteria(value))
		}
	}

	// Keyset setting
	if params.TokenUUID != "" {
		if strings.ToLower(filterMap["timestamp"]) == "true" {
			statement += AndCriteria(fmt.Sprintf(`UPDATE_TIME <= (SELECT UPDATE_TIME FROM MEDIA WHERE EXTERNAL_ID = '%s' AND DELETED = FALSE)`,
				params.TokenUUID))
			// Add footer
			statement += fmt.Sprintf(`DELETED = FALSE
			ORDER BY UPDATE_TIME DESC
			FETCH FIRST %d ROWS ONLY;`, params.Size)
		} else {
			statement += AndCriteria(fmt.Sprintf(`MEDIA_ID >= (SELECT MEDIA_ID FROM MEDIA WHERE EXTERNAL_ID = '%s' AND DELETED = FALSE)`,
				params.TokenUUID))

			// Add footer
			statement += fmt.Sprintf(`DELETED = FALSE
			ORDER BY MEDIA_ID ASC
			FETCH FIRST %d ROWS ONLY;`, params.Size)
		}
	} else if params.TokenID > 0 {
		// Using default ID
		if filterMap["timestamp"] == "true" {
			statement += AndCriteria(fmt.Sprintf(`UPDATE_TIME <= (SELECT UPDATE_TIME FROM MEDIA WHERE MEDIA_ID = %d AND DELETED = FALSE)`,
				params.TokenID))

			statement += fmt.Sprintf(`DELETED = FALSE
			ORDER BY UPDATE_TIME DESC
			FETCH FIRST %d ROWS ONLY;`, params.Size)
		} else {
			statement += AndCriteria(fmt.Sprintf(`MEDIA_ID >= %d`,
				params.TokenID))

			// Add footer
			statement += fmt.Sprintf(`DELETED = FALSE
			ORDER BY MEDIA_ID ASC
			FETCH FIRST %d ROWS ONLY;`, params.Size)
		}

	} else {
		// By timestamp
		// Add footer
		statement += fmt.Sprintf(`DELETED = FALSE
			ORDER BY UPDATE_TIME DESC
			FETCH FIRST %d ROWS ONLY;`, params.Size)
	}

	rows, err := conn.QueryContext(m.ctx, statement)
	if rows != nil && rows.Err() != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		err = rows.Close()
	}()

	medias := make([]*domain.MediaAggregate, 0)
	for rows.Next() {
		media := new(domain.MediaAggregate)
		err = rows.Scan(&media.MediaID, &media.ExternalID, &media.Title, &media.DisplayName, &media.Description, &media.UserID, &media.AuthorID,
			&media.PublishDate, &media.MediaType, &media.CreateTime, &media.UpdateTime, &media.DeleteTime, &media.Metadata, &media.Deleted)
		if err != nil {
			return nil, err
		}
		medias = append(medias, media)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(medias) == 0 {
		return nil, fmt.Errorf("%w", global.EntitiesNotFound)
	}

	return medias, nil
}

func (m *MediaRDBMSRepository) FetchByID(id int64, external_id string) (*domain.MediaAggregate, error) {
	conn, err := m.db.Conn(m.ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = conn.Close()
	}()

	media := new(domain.MediaAggregate)
	var row *sql.Row
	if id > 0 {
		statement := `SELECT * FROM MEDIA WHERE MEDIA_ID = $1 AND DELETED = FALSE`
		row = conn.QueryRowContext(m.ctx, statement, id)
	} else {
		statement := `SELECT * FROM MEDIA WHERE EXTERNAL_ID = $1 AND DELETED = FALSE`
		row = conn.QueryRowContext(m.ctx, statement, external_id)
	}

	if row == nil {
		return nil, fmt.Errorf("%w", global.EntityNotFound)
	}

	err = row.Scan(&media.MediaID, &media.ExternalID, &media.Title, &media.DisplayName, &media.Description, &media.UserID, &media.AuthorID,
		&media.PublishDate, &media.MediaType, &media.CreateTime, &media.UpdateTime, &media.DeleteTime, &media.Metadata, &media.Deleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w", global.EntityNotFound)
		}

		return nil, err
	}

	return media, nil
}

func (m *MediaRDBMSRepository) FetchByTitle(title string) (*domain.MediaAggregate, error) {
	conn, err := m.db.Conn(m.ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = conn.Close()
	}()

	statement := `SELECT * FROM MEDIA WHERE LOWER(TITLE) = LOWER($1) AND DELETED = FALSE`

	media := new(domain.MediaAggregate)
	row := conn.QueryRowContext(m.ctx, statement, title)
	if row == nil {
		return nil, fmt.Errorf("%w", global.EntityNotFound)
	}

	err = row.Scan(&media.MediaID, &media.ExternalID, &media.Title, &media.DisplayName, &media.Description, &media.UserID, &media.AuthorID,
		&media.PublishDate, &media.MediaType, &media.CreateTime, &media.UpdateTime, &media.DeleteTime, &media.Metadata, &media.Deleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w", global.EntityNotFound)
		}

		return nil, err
	}

	return media, nil
}

func (m *MediaRDBMSRepository) UpdateOne(id int64, external_id string, mediaUpdated *domain.MediaAggregate) error {
	conn, err := m.db.Conn(m.ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()

	mediaUpdated.UpdateTime = time.Now()

	if id > 0 {
		statement := `UPDATE MEDIA SET TITLE = $1, DISPLAY_NAME = $2, DESCRIPTION = $3, USER_ID = $4, AUTHOR_ID = $5, PUBLISH_DATE = $6, MEDIA_TYPE = $7, 
                 UPDATE_TIME = $8 WHERE MEDIA_ID = $9`

		_, err = conn.ExecContext(m.ctx, statement, mediaUpdated.Title, mediaUpdated.DisplayName, mediaUpdated.Description, mediaUpdated.UserID, mediaUpdated.AuthorID,
			mediaUpdated.PublishDate, mediaUpdated.MediaType, mediaUpdated.UpdateTime, id)
	} else {
		statement := `UPDATE MEDIA SET TITLE = $1, DISPLAY_NAME = $2, DESCRIPTION = $3, USER_ID = $4, AUTHOR_ID = $5, PUBLISH_DATE = $6, MEDIA_TYPE = $7, 
                 UPDATE_TIME = $8 WHERE EXTERNAL_ID = $9`

		_, err = conn.ExecContext(m.ctx, statement, mediaUpdated.Title, mediaUpdated.DisplayName, mediaUpdated.Description, mediaUpdated.UserID, mediaUpdated.AuthorID,
			mediaUpdated.PublishDate, mediaUpdated.MediaType, mediaUpdated.UpdateTime, external_id)
	}

	if customErr, ok := err.(*pq.Error); ok {
		if customErr.Code == "23505" {
			return global.EntityExists
		}
	}

	return err
}

func (m *MediaRDBMSRepository) RemoveOne(id int64, external_id string) error {
	conn, err := m.db.Conn(m.ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()

	// statement := `DELETE FROM MEDIA WHERE MEDIA_ID = $1`

	if id > 0 {
		statement := `UPDATE MEDIA SET DELETED = TRUE WHERE MEDIA_ID = $1`
		_, err = conn.ExecContext(m.ctx, statement, id)
	} else {
		statement := `UPDATE MEDIA SET DELETED = TRUE WHERE EXTERNAL_ID = $1`
		_, err = conn.ExecContext(m.ctx, statement, external_id)
	}

	return err
}
