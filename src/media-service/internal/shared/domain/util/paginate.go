package util

import (
	"github.com/google/uuid"
	"strconv"
)

// PaginationParams Parameters required for database pagination
type PaginationParams struct {
	TokenID   int64
	TokenUUID string
	Size      int32
}

func NewPaginationParams(TokenIDString, TokenUUID, sizeString string) *PaginationParams {
	var pageTokenID, size int64
	var err error

	if TokenIDString != "" {
		pageTokenID, err = strconv.ParseInt(TokenIDString, 10, 64)
		if err != nil {
			pageTokenID = 1
		}
	} else {
		pageTokenID = 1
	}

	if sizeString != "" {
		size, err = strconv.ParseInt(sizeString, 10, 32)
		if err != nil {
			size = 10
		}
	} else {
		size = 10
	}

	params := &PaginationParams{
		TokenID:   pageTokenID,
		TokenUUID: TokenUUID,
		Size:      int32(size),
	}
	params.Sanitize()

	return params
}

func (p *PaginationParams) Sanitize() {
	if p.TokenID <= 0 {
		p.TokenID = 1
	} else if p.Size > 100 {
		p.Size = 100
	} else if p.Size <= 0 {
		p.Size = 10
	} else if p.TokenUUID != "" {
		_, err := uuid.Parse(p.TokenUUID)
		if err != nil {
			p.TokenUUID = ""
		}
	}
}

func GetIndex(pageTokenID int64, size int32) int64 {
	// Index-from-limit algorithm formula
	// f(x)= w-x
	// w (omega) = xy
	// where x = limit and y = page
	return int64(size)*pageTokenID - int64(size)
}
