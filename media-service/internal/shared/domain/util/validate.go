package util

import (
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/global"
	"strconv"
)

func SanitizeID(id string) (int64, error) {
	idSanitized, err := strconv.ParseInt(id, 10, 64)
	if err != nil || idSanitized <= 0 {
		return 0, global.InvalidID
	}

	return idSanitized, nil
}

func SanitizeUUID(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return global.InvalidID
	}

	return nil
}
