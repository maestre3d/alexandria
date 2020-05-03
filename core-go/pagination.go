package core

import (
	"strconv"
)

// PaginationParams Required struct to handle fetch pagination
type PaginationParams struct {
	Token string
	Size  int
}

// NewPaginationParams Create a new pagination token and size, uses UUID v1 as Token/ID
func NewPaginationParams(token, size string) *PaginationParams {
	// Pagination default values
	params := &PaginationParams{
		Token: "",
		Size:  10,
	}

	// If fields are valid, then change default param's values
	err := ValidateUUID(token)
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
