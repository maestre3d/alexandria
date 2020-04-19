package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/lib/pq"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/exception"
	"strings"
	"time"

	"github.com/maestre3d/alexandria/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"
)

type MediaDBMSRepository struct {
	db     *sql.DB
	mem    *redis.Client
	logger log.Logger
	ctx    context.Context
}

func NewMediaRDBMSRepository(db *sql.DB, mem *redis.Client, logger log.Logger, ctx context.Context) *MediaRDBMSRepository {
	return &MediaRDBMSRepository{
		db,
		mem,
		logger,
		ctx,
	}
}

func (m *MediaRDBMSRepository) Save(media *domain.MediaEntity) error {
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
			return exception.EntityExists
		}
	}

	return err
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
			return exception.EntityExists
		}
	}

	if err != nil {
		return err
	}

	// Async Write-through pattern
	if m.mem != nil {
		go func() {
			memConn := m.mem.Conn()
			defer func() {
				err = memConn.Close()
			}()

			err := memConn.Get(fmt.Sprintf(`media:%s`, external_id)).Err()
			if err == nil {
				mediaJSON, err := json.Marshal(mediaUpdated)
				if err == nil {
					err = memConn.Set(fmt.Sprintf(`media:%s`, external_id), mediaJSON, (time.Hour * time.Duration(24))).Err()
				}
			}
		}()
	}

	return nil
}

func (m *MediaRDBMSRepository) RemoveOne(id int64, external_id string) error {
	conn, err := m.db.Conn(m.ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()

	// Soft-delete
	statement := `UPDATE MEDIA SET DELETED = TRUE WHERE EXTERNAL_ID = $1 AND DELETED = FALSE`
	_, err = conn.ExecContext(m.ctx, statement, external_id)

	// Async Write-through pattern
	if m.mem != nil {
		go func() {
			memConn := m.mem.Conn()
			defer func() {
				err = memConn.Close()
			}()

			err := memConn.Get(fmt.Sprintf(`media:%s`, external_id)).Err()
			if err == nil {
				err = memConn.Del(fmt.Sprintf(`media:%s`, external_id)).Err()
			}
		}()
	}

	return err
}

func (m *MediaRDBMSRepository) FetchByID(id string) (*domain.MediaEntity, error) {
	conn, err := m.db.Conn(m.ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = conn.Close()
	}()

	// Cache-aside pattern first
	if m.mem != nil {
		mediaChan := make(chan *domain.MediaEntity)

		go func() {
			memConn := m.mem.Conn()
			defer func() {
				err := memConn.Close()
				if err != nil {
					mediaChan <- nil
				}
			}()

			mediaMem, err := memConn.Get(fmt.Sprintf(`media:%s`, id)).Result()
			if err == nil {
				media := new(domain.MediaEntity)
				err = json.Unmarshal([]byte(mediaMem), media)
				if err == nil {
					mediaChan <- media
				}

				mediaChan <- nil
			}
			mediaChan <- nil
		}()

		select {
		case media := <-mediaChan:
			if media != nil {
				return media, nil
			}
		}

		close(mediaChan)
	}

	statement := `SELECT * FROM MEDIA WHERE EXTERNAL_ID = $1 AND DELETED = FALSE`

	media := new(domain.MediaEntity)
	err = conn.QueryRowContext(m.ctx, statement, id).Scan(&media.MediaID, &media.ExternalID, &media.Title, &media.DisplayName, &media.Description, &media.UserID, &media.AuthorID,
		&media.PublishDate, &media.MediaType, &media.CreateTime, &media.UpdateTime, &media.DeleteTime, &media.Metadata, &media.Deleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	// Async Write-through pattern
	if m.mem != nil {
		go func() {
			memConn := m.mem.Conn()
			defer func() {
				err = memConn.Close()
			}()

			mediaJSON, err := json.Marshal(media)
			if err == nil {
				err = memConn.Set(fmt.Sprintf(`media:%s`, media.ExternalID), mediaJSON, (time.Hour * time.Duration(24))).Err()
			}
		}()
	}

	return media, nil
}

func (m *MediaRDBMSRepository) Fetch(params *util.PaginationParams, filterParams util.FilterParams) ([]*domain.MediaEntity, error) {
	conn, err := m.db.Conn(m.ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = conn.Close()
	}()

	if params == nil {
		params = util.NewPaginationParams("", "10")
	}

	// Using keyset pagination along with page_tokens (ref. Google API Design)
	statement := `SELECT * FROM MEDIA WHERE `
	params.Size += 1

	// Criteria map filter -> Query Builder
	for filterKey, value := range filterParams {
		switch {
		case filterKey == "query" && value != "":
			statement += AndCriteriaSQL(QueryCriteriaSQL(value))
		case filterKey == "author" && value != "":
			statement += AndCriteriaSQL(AuthorCriteriaSQL(value))
		case filterKey == "user" && value != "":
			statement += AndCriteriaSQL(PublisherCriteriaSQL(value))
		case filterKey == "media" && value != "":
			statement += AndCriteriaSQL(MediaTypeCriteriaSQL(value))
		}
	}

	// Keyset
	if params.Token != "" {
		if filterParams["timestamp"] == "false" {
			statement += AndCriteriaSQL(fmt.Sprintf(`ID >= (SELECT MEDIA_ID FROM MEDIA WHERE EXTERNAL_ID = '%s' AND DELETED = FALSE)`,
				params.Token))
			statement += fmt.Sprintf(`DELETED = FALSE ORDER BY ID ASC FETCH FIRST %d ROWS ONLY`, params.Size)
		} else if filterParams["timestamp"] == "" || filterParams["timestamp"] == "true" {
			// Timestamp/Most recent by default
			statement += AndCriteriaSQL(fmt.Sprintf(`UPDATE_TIME <= (SELECT UPDATE_TIME FROM MEDIA WHERE EXTERNAL_ID = '%s' AND DELETED = FALSE)`,
				params.Token))
			statement += fmt.Sprintf(`DELETED = FALSE ORDER BY UPDATE_TIME DESC FETCH FIRST %d ROWS ONLY`, params.Size)
		}
	} else {
		if filterParams["timestamp"] == "false" {
			statement += fmt.Sprintf(`DELETED = FALSE ORDER BY ID ASC FETCH FIRST %d ROWS ONLY`, params.Size)
		} else if filterParams["timestamp"] == "" || filterParams["timestamp"] == "true" {
			// Timestamp/Most recent by default
			statement += fmt.Sprintf(`DELETED = FALSE ORDER BY UPDATE_TIME DESC FETCH FIRST %d ROWS ONLY`, params.Size)
		}
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

	medias := make([]*domain.MediaEntity, 0)
	for rows.Next() {
		media := new(domain.MediaEntity)
		err = rows.Scan(&media.MediaID, &media.ExternalID, &media.Title, &media.DisplayName, &media.Description, &media.UserID, &media.AuthorID,
			&media.PublishDate, &media.MediaType, &media.CreateTime, &media.UpdateTime, &media.DeleteTime, &media.Metadata, &media.Deleted)
		if err != nil {
			continue
		}
		medias = append(medias, media)
	}

	return medias, nil
}
