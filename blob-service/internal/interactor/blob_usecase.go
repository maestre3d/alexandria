package interactor

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	"strconv"
	"strings"
)

type Blob struct {
	logger     log.Logger
	repository domain.BlobRepository
	storage    domain.BlobStorage
	eventBus   domain.BlobEvent
}

func NewBlob(logger log.Logger, repo domain.BlobRepository, storage domain.BlobStorage, eventBus domain.BlobEvent) *Blob {
	return &Blob{
		logger:     logger,
		repository: repo,
		storage:    storage,
		eventBus:   eventBus,
	}
}

func (u *Blob) Store(ctx context.Context, ag *domain.BlobAggregate) (*domain.Blob, error) {
	size, err := strconv.ParseInt(ag.Size, 10, 64)
	if err != nil {
		return nil, exception.NewErrorDescription(exception.InvalidFieldFormat,
			fmt.Sprintf(exception.InvalidFieldFormatString, "size", "int64/bigint"))
	}

	blob := domain.NewBlob(ag.RootID, ag.Service, ag.BlobType, ag.Extension, size)
	err = blob.IsValid()
	if err != nil {
		return nil, err
	}

	blob.Content = ag.Content
	defer blob.Content.Close()

	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()
	// Storage is our priority, persistence is just for logging/admin tasks
	err = u.storage.Store(ctxR, blob)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Rollback
		if err != nil {
			if errRoll := u.storage.Delete(ctxR, blob.Name, blob.Service); errRoll != nil {
				_ = u.logger.Log("method", "blob.interactor.store", "err", errRoll.Error())
			}
		}
	}()

	err = u.repository.Save(ctxR, *blob)
	if err != nil {
		return nil, err
	}

	// Set root_id (entity) so the client won't be confused,
	// avoid using service_id prefix when returning to client
	// since it's just an internal implementation.
	// It's more semantic if we keep using root_id as default ID
	blob.ID = ag.RootID

	// Start transaction
	errC := make(chan error)
	go func() {
		ctxE, cancelE := context.WithCancel(ctx)
		defer cancelE()
		errC <- u.eventBus.Uploaded(ctxE, *blob)
	}()

	select {
	case err = <-errC:
		if err != nil {
			_ = u.logger.Log("method", "blob.interactor.store", "msg",
				fmt.Sprintf("%s_%s event sending failed", strings.ToUpper(blob.Service), domain.BlobUploaded),
				"err", err.Error())
			return nil, err
		}
		_ = u.logger.Log("method", "blob.interactor.store", "msg",
			fmt.Sprintf("%s_%s integration event published", strings.ToUpper(blob.Service), domain.BlobUploaded))
		break
	}

	return blob, nil
}

func (u *Blob) Get(ctx context.Context, id, service string) (*domain.Blob, error) {
	prefID := domain.GetServiceID(service) + id

	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()
	blob, err := u.repository.FetchByID(ctxR, prefID)
	if err != nil {
		return nil, err
	}

	blob.ID = id
	return blob, nil
}

func (u *Blob) Delete(ctx context.Context, id, service string) error {
	prefID := domain.GetServiceID(service) + id

	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()
	err := u.storage.Delete(ctxR, id, service)
	if err != nil {
		return err
	}

	err = u.repository.Remove(ctxR, prefID)
	if err != nil {
		u.logger.Log("method", "blob.interactor.store", "msg",
			"persistence removal failed",
			"err", err.Error())
	}

	errC := make(chan error)
	go func() {
		ctxE, cancelE := context.WithCancel(ctx)
		defer cancelE()
		errC <- u.eventBus.Removed(ctxE, id, service)
	}()

	select {
	case err = <-errC:
		if err != nil {
			_ = u.logger.Log("method", "blob.interactor.store", "msg",
				fmt.Sprintf("%s_%s event sending failed", strings.ToUpper(service), domain.BlobUploaded),
				"err", err.Error())
			return err
		}
		_ = u.logger.Log("method", "blob.interactor.delete", "msg",
			fmt.Sprintf("%s_%s event published", strings.ToUpper(service), domain.BlobRemoved))
		break
	}

	return nil
}
