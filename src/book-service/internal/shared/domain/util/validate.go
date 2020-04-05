package util

import (
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/global"
	"strconv"
)

func ValidateID(id string) (int64, error) {
	idSanitized, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, err
	} else if idSanitized <= 0 {
		return 0, global.InvalidID
	}

	return idSanitized, nil
}
