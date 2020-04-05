package global

import "strconv"

// PaginationParams Parameters required for database pagination
type PaginationParams struct {
	Page  int64
	Limit int64
}

func NewPaginationParams(pageString, limitString string) *PaginationParams {
	var page, limit int64
	var err error

	if pageString != "" {
		page, err = strconv.ParseInt(pageString, 10, 64)
		if err != nil {
			page = 1
		}
	} else {
		page = 1
	}

	if limitString != "" {
		limit, err = strconv.ParseInt(limitString, 10, 64)
		if err != nil {
			limit = 10
		}
	} else {
		limit = 10
	}

	params := &PaginationParams{
		Page:  page,
		Limit: limit,
	}
	params.Sanitize()

	return params
}

func (p *PaginationParams) Sanitize() {
	if p.Page <= 0 {
		p.Page = 1
	} else if p.Limit > 100 {
		p.Limit = 100
	} else if p.Limit <= 0 {
		p.Limit = 10
	}
}
