package domain

import "context"

type UserRepository interface {
	FetchByID(ctx context.Context, id string) (*User, error)
	ReplacePicture(ctx context.Context, id, pictureURL string) error
}
