package util

import (
	"strconv"

	"github.com/google/uuid"
)

// FilterParams Optional fields for query filtering
type FilterParams map[string]string

// PaginationParams Required struct to handle fetch pagination
type PaginationParams struct {
	Token string
	Size  int
}

// NewPaginationParams Create a new pagination token and size
func NewPaginationParams(token, size string) *PaginationParams {
	// Pagination default values
	params := &PaginationParams{
		Token: "",
		Size:  10,
	}

	// If fields are valid, then change default param's values
	_, err := uuid.Parse(token)
	if err == nil {
		params.Token = token
	}

	sizeInt, err := strconv.Atoi(size)
	if err == nil && sizeInt > 0 {
		if sizeInt > 100 {
			params.Size = 100
		} else {
			params.Size = sizeInt
		}
	}

	return params
}
