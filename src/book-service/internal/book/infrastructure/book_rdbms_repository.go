package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/maestre3d/alexandria/src/book-service/internal/book/domain"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/util"
)

type BookRDBMSRepository struct {
	db     *sql.DB
	logger util.ILogger
	ctx    context.Context
}

func NewBookRDBMSRepository(db *sql.DB, logger util.ILogger, ctx context.Context) *BookRDBMSRepository {
	return &BookRDBMSRepository{
		db,
		logger,
		ctx,
	}
}

func (b *BookRDBMSRepository) Save(book *domain.BookEntity) error {
	conn, err := b.db.Conn(b.ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	statement := `INSERT INTO BOOK (TITLE, PUBLISHED_AT, CREATED_AT, UPDATED_AT, UPLOADED_BY, AUTHOR) VALUES
	($1, $2, $3, $4, $5, $6);`

	_, err = conn.ExecContext(b.ctx, statement, book.Title, book.PublishedAt, book.CreatedAt, book.UpdatedAt, book.UploadedBy, book.Author)

	return err
}

func (b *BookRDBMSRepository) Fetch(params *global.PaginationParams) ([]*domain.BookEntity, error) {
	conn, err := b.db.Conn(b.ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	index := util.GetIndex(params.Page, params.Limit)
	statement := fmt.Sprintf(`SELECT * FROM BOOK WHERE ID > %d ORDER BY ID ASC FETCH FIRST %d ROWS ONLY`, index, params.Limit)

	rows, err := conn.QueryContext(b.ctx, statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := make([]*domain.BookEntity, 0)
	for rows.Next() {
		book := new(domain.BookEntity)
		err := rows.Scan(&book.ID, &book.Title, &book.PublishedAt, &book.CreatedAt, &book.UpdatedAt, &book.UploadedBy, &book.Author)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}

func (b *BookRDBMSRepository) FetchByID(id int64) (*domain.BookEntity, error) {
	conn, err := b.db.Conn(b.ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	statement := `SELECT * FROM BOOK WHERE ID = $1`

	book := new(domain.BookEntity)
	result := conn.QueryRowContext(b.ctx, fmt.Sprintf(statement, id)).Scan(&book.ID, &book.Title, &book.PublishedAt, &book.CreatedAt, &book.UpdatedAt, &book.UploadedBy, &book.Author)
	if result.Error() != "" {
		return nil, errors.New(result.Error())
	}

	return book, nil
}

func (b *BookRDBMSRepository) FetchByTitle(title string) (*domain.BookEntity, error) {
	conn, err := b.db.Conn(b.ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	statement := `SELECT * FROM BOOK WHERE LOWER(TITLE) = LOWER($1)`

	book := new(domain.BookEntity)
	result := conn.QueryRowContext(b.ctx, statement, title).Scan(&book.ID, &book.Title, &book.PublishedAt, &book.CreatedAt, &book.UpdatedAt, &book.UploadedBy, &book.Author)
	if result.Error() != "" {
		if result.Error() == "sql: no rows in result set" {
			return nil, nil
		}

		return nil, errors.New(result.Error())
	}

	return book, nil
}

func (b *BookRDBMSRepository) UpdateOne(id int64, bookUpdated *domain.BookEntity) error {
	conn, err := b.db.Conn(b.ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	statement := `UPDATE BOOK SET TITLE = $1, PUBLISHED_AT = $2, UPDATED_AT = $3, AUTHOR = $4 WHERE ID = $5`

	_, err = conn.ExecContext(b.ctx, fmt.Sprintf(statement, bookUpdated.Title, bookUpdated.PublishedAt, bookUpdated.UpdatedAt, bookUpdated.Author, id))
	return err
}

func (b *BookRDBMSRepository) RemoveOne(id int64) error {
	conn, err := b.db.Conn(b.ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	statement := `DELETE FROM BOOK WHERE ID = $1`

	_, err = conn.ExecContext(b.ctx, fmt.Sprintf(statement, id))
	return err
}
