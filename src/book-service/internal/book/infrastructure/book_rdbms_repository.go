package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/multierr"

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
	defer func() {
		err = multierr.Append(err, conn.Close())
	}()

	statement := `INSERT INTO BOOK (TITLE, PUBLISHED_AT, CREATED_AT, UPDATED_AT, UPLOADED_BY, AUTHOR) VALUES
	($1, $2, $3, $4, $5, $6);`

	_, errExec := conn.ExecContext(b.ctx, statement, book.Title, book.PublishedAt, book.CreatedAt, book.UpdatedAt, book.UploadedBy, book.Author)
	err = multierr.Append(err, errExec)

	return err
}

func (b *BookRDBMSRepository) Fetch(params *global.PaginationParams) ([]*domain.BookEntity, error) {
	conn, err := b.db.Conn(b.ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = multierr.Append(err, conn.Close())
	}()

	index := util.GetIndex(params.Page, params.Limit)
	statement := fmt.Sprintf(`SELECT * FROM BOOK WHERE ID > %d ORDER BY ID ASC FETCH FIRST %d ROWS ONLY`, index, params.Limit)

	rows, errQuery := conn.QueryContext(b.ctx, statement)
	if rows != nil && rows.Err() != nil {
		return nil, multierr.Append(err, rows.Err())
	}
	if errQuery != nil {
		return nil, multierr.Append(err, errQuery)
	}
	defer func() {
		err = multierr.Append(err, rows.Close())
	}()

	books := make([]*domain.BookEntity, 0)
	for rows.Next() {
		book := new(domain.BookEntity)
		errScan := rows.Scan(&book.ID, &book.Title, &book.PublishedAt, &book.CreatedAt, &book.UpdatedAt, &book.UploadedBy, &book.Author)
		if errScan != nil {
			return nil, multierr.Append(err, errScan)
		}
		books = append(books, book)
	}

	if errRow := rows.Err(); errRow != nil {
		return nil, multierr.Append(err, errRow)
	}

	if len(books) == 0 {
		return nil, multierr.Append(err, global.EntitiesNotFound)
	}

	return books, nil
}

func (b *BookRDBMSRepository) FetchByID(id int64) (*domain.BookEntity, error) {
	conn, err := b.db.Conn(b.ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = multierr.Append(err, conn.Close())
	}()

	statement := `SELECT * FROM BOOK WHERE ID = $1`

	book := new(domain.BookEntity)
	row := conn.QueryRowContext(b.ctx, statement, id)
	if row == nil {
		return nil, multierr.Append(err, global.EntityNotFound)
	}

	errQuery := row.Scan(&book.ID, &book.Title, &book.PublishedAt, &book.CreatedAt, &book.UpdatedAt, &book.UploadedBy, &book.Author)
	if errQuery != nil {
		if errQuery == sql.ErrNoRows {
			return nil, multierr.Append(err, global.EntityNotFound)
		}

		return nil, multierr.Append(err, errQuery)
	}

	return book, nil
}

func (b *BookRDBMSRepository) FetchByTitle(title string) (*domain.BookEntity, error) {
	conn, err := b.db.Conn(b.ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = multierr.Append(err, conn.Close())
	}()

	statement := `SELECT * FROM BOOK WHERE LOWER(TITLE) = LOWER($1)`

	book := new(domain.BookEntity)
	row := conn.QueryRowContext(b.ctx, statement, title)
	if row == nil {
		return nil, multierr.Append(err, global.EntityNotFound)
	}

	errQuery := row.Scan(&book.ID, &book.Title, &book.PublishedAt, &book.CreatedAt, &book.UpdatedAt, &book.UploadedBy, &book.Author)
	if errQuery != nil {
		if errQuery == sql.ErrNoRows {
			return nil, multierr.Append(err, global.EntityNotFound)
		}

		return nil, multierr.Append(err, errQuery)
	}

	return book, nil
}

func (b *BookRDBMSRepository) UpdateOne(id int64, bookUpdated *domain.BookEntity) error {
	conn, err := b.db.Conn(b.ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = multierr.Append(err, conn.Close())
	}()

	statement := `UPDATE BOOK SET TITLE = $1, PUBLISHED_AT = $2, UPDATED_AT = $3, AUTHOR = $4 WHERE ID = $5`

	_, errExec := conn.ExecContext(b.ctx, fmt.Sprintf(statement, bookUpdated.Title, bookUpdated.PublishedAt, bookUpdated.UpdatedAt, bookUpdated.Author, id))
	return multierr.Append(err, errExec)
}

func (b *BookRDBMSRepository) RemoveOne(id int64) error {
	conn, err := b.db.Conn(b.ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = multierr.Append(err, conn.Close())
	}()

	statement := `DELETE FROM BOOK WHERE ID = $1`

	_, errExec := conn.ExecContext(b.ctx, fmt.Sprintf(statement, id))
	return multierr.Append(err, errExec)
}
