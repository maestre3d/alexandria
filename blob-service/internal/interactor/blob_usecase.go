package interactor

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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

	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	operationKind := "update"
	prefID := domain.GetServiceID(ag.Service) + ag.RootID
	blob, err := u.repository.FetchByID(ctxR, prefID)
	snapshot := blob
	if err != nil {
		// If not exists, create new blob entity
		blob = domain.NewBlob(ag.RootID, ag.Service, ag.BlobType, ag.Extension, size)
		err = blob.IsValid()
		if err != nil {
			return nil, err
		}
		operationKind = "create"
	}

	blob.Content = ag.Content
	defer blob.Content.Close()

	// Storage is our priority
	err = u.storage.Store(ctxR, blob)
	if err != nil {
		return nil, err
	}

	// If any single error, then rollback persistence
	defer func() {
		// Rollback
		if err != nil {
			if operationKind == "create" {
				if errRoll := u.storage.Delete(ctxR, blob.Name, blob.Service); errRoll != nil {
					_ = u.logger.Log("method", "blob.interactor.store", "err", errRoll.Error())
				}
			} else {
				err = u.repository.Save(ctxR, *snapshot)
			}

			if err != nil {
				_ = u.logger.Log("method", "blob.interactor.store", "err", err.Error())
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
		errC <- u.eventBus.Uploaded(ctxE, *blob, snapshot)
	}()

	select {
	case err = <-errC:
		if err != nil {
			// Rollback persistence
			if operationKind == "create" {
				errR := u.repository.Remove(ctxR, prefID)
				_ = level.Error(u.logger).Log("err", errR)
			}
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

	// Retrieve blob entity from database to get file extension
	blob, err := u.repository.FetchByID(ctxR, prefID)
	if err != nil {
		return err
	}

	err = u.storage.Delete(ctxR, blob.Name, blob.Service)
	if err != nil {
		_ = u.logger.Log("method", "blob.interactor.delete", "msg",
			"blob storage removal failed",
			"err", err.Error())
		return err
	}

	err = u.repository.Remove(ctxR, prefID)
	if err != nil {
		_ = u.logger.Log("method", "blob.interactor.delete", "msg",
			"persistence removal failed",
			"err", err.Error())
		return err
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
			_ = u.logger.Log("method", "blob.interactor.delete", "msg",
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
