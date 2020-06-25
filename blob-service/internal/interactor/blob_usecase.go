package interactor

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	"mime/multipart"
	"strings"
)

type Blob struct {
	logger     log.Logger
	repository domain.BlobRepository
	storage    domain.BlobStorage
}

func NewBlob(logger log.Logger, repo domain.BlobRepository, storage domain.BlobStorage) *Blob {
	return &Blob{
		logger:     logger,
		repository: repo,
		storage:    storage,
	}
}

func (u *Blob) Store(ctx context.Context, rootID, service string, header *multipart.FileHeader) (*domain.Blob, error) {
	contentSlice := strings.Split(header.Header.Get("Content-Type"), "/")
	if len(contentSlice) <= 1 {
		return nil, exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"file", "invalid file extension"))
	}

	blob := domain.NewBlob(rootID, service, contentSlice[0], contentSlice[1], header.Size)
	err := blob.IsValid()
	if err != nil {
		return nil, err
	}

	blob.File = header

	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()
	// Storage is our priority, persistence is just for logging
	err = u.storage.Store(ctxR, blob)
	if err != nil {
		return nil, err
	}

	err = u.repository.Save(ctxR, *blob)
	if err != nil {
		// Rollback
		if errRoll := u.storage.Delete(ctxR, blob.Name, blob.Service); errRoll != nil {
			_ = u.logger.Log("method", "blob.interactor.store", "err", errRoll.Error())
		}
		return nil, err
	}

	return blob, nil
}

func (u *Blob) Get(ctx context.Context, id string) (*domain.Blob, error) {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()
	return u.repository.FetchByID(ctxR, id)
}

func (u *Blob) Delete(ctx context.Context, id string) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	blob, err := u.repository.FetchByID(ctxR, id)
	if err != nil {
		return err
	}

	err = u.storage.Delete(ctxR, blob.Name, blob.Service)
	if err != nil {
		return err
	}

	return u.repository.Remove(ctxR, id)
}
