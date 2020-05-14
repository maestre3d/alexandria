package domain

import "github.com/alexandria-oss/core"

type UserRepository interface {
	Save(user User) error
	Fetch(params core.PaginationParams, filter core.FilterParams) ([]*User, error)
	FetchByID(id string) (*User, error)
	Replace(user User) error
	Remove(id string) error
}
