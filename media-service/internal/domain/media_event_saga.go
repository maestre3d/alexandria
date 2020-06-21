package domain

import (
	"context"
)

const (
	MediaCreated   = "MEDIA_CREATED"
	OwnerVerify    = "OWNER_VERIFY"
	OwnerVerified  = "MEDIA_OWNER_VERIFIED"
	OwnerFailed    = "MEDIA_OWNER_FAILED"
	AuthorVerify   = "AUTHOR_VERIFY"
	AuthorVerified = "MEDIA_AUTHOR_VERIFIED"
	AuthorFailed   = "MEDIA_AUTHOR_FAILED"
)

type MediaEventSAGA interface {
	VerifyAuthor(ctx context.Context, authors []string) error
	Created(ctx context.Context, media Media) error
}
