package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	"gocloud.dev/blob"
	"io"
	"sync"
)

type BlobS3Storage struct {
	logger log.Logger
	mu     *sync.Mutex
}

func NewBlobS3Storage(logger log.Logger) *BlobS3Storage {
	return &BlobS3Storage{
		logger: logger,
		mu:     new(sync.Mutex),
	}
}

func (s *BlobS3Storage) Store(ctx context.Context, blobRef *domain.Blob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, err := blob.OpenBucket(ctx, fmt.Sprintf("s3://%s?region=%s", domain.StorageDomain, domain.StorageRegion))
	if err != nil {
		return err
	}
	defer bucket.Close()
	bucket = blob.PrefixedBucket(bucket, domain.StoragePath+"/"+blobRef.Service+"/")

	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()
	w, err := bucket.NewWriter(ctxR, blobRef.Name, nil)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, blobRef.Content)
	if err != nil {
		return err
	}

	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	closeErr := w.Close()
	if closeErr != nil {
		return closeErr
	}

	return nil
}

func (s *BlobS3Storage) Delete(ctx context.Context, key, service string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, err := blob.OpenBucket(ctx, fmt.Sprintf("s3://%s?region=%s", domain.StorageDomain, domain.StorageRegion))
	if err != nil {
		return err
	}
	defer bucket.Close()
	bucket = blob.PrefixedBucket(bucket, "alexandria/"+service+"/")

	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()
	return bucket.Delete(ctxR, key)
}
