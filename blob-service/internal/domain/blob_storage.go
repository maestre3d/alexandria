package domain

import "context"

type BlobStorage interface {
	Store(ctx context.Context, blobRef *Blob) error
	Delete(ctx context.Context, key, service string) error
}
