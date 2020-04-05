package infrastructure

import (
	"database/sql"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/util"
)

type BookRDBMSRepository struct {
	conn   *sql.Conn
	logger util.ILogger
}
