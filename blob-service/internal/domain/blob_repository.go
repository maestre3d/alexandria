package domain

import "context"

type BlobRepository interface {
	Save(ctx context.Context, blobRef Blob) error
	FetchByID(ctx context.Context, id string) (*Blob, error)
	Remove(ctx context.Context, id string) error
}
