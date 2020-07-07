package interactor

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
)

type BlobSAGA struct {
	logger  log.Logger
	repo    domain.BlobRepository
	storage domain.BlobStorage
}

func NewBlobSaga(logger log.Logger, repo domain.BlobRepository, storage domain.BlobStorage) *BlobSAGA {
	return &BlobSAGA{
		logger:  logger,
		repo:    repo,
		storage: storage,
	}
}

func (u *BlobSAGA) Failed(ctx context.Context, rootID, service string, snapshotJSON []byte) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	var snapshot *domain.Blob
	prefID := domain.GetServiceID(service) + rootID
	blob, err := u.repo.FetchByID(ctxR, prefID)
	defer func() {
		if err != nil {
			_ = level.Error(u.logger).Log("err", err)
		}
	}()
	if err != nil {
		return err
	}

	err = json.Unmarshal(snapshotJSON, &snapshot)
	if err != nil {
		// If not a valid snapshot, then rollback create operation
		err = u.storage.Delete(ctxR, blob.Name, service)
		if err != nil {
			return err
		}
		err = u.repo.Remove(ctxR, prefID)
		if err != nil {
			return err
		}
		return nil
	}

	// Rollback update operation
	err = u.repo.Save(ctxR, *snapshot)
	if err != nil {
		return err
	}

	return nil
}
