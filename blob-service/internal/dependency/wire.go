// +build wireinject

package dependency

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	"github.com/maestre3d/alexandria/blob-service/internal/infrastructure"
	"github.com/maestre3d/alexandria/blob-service/internal/interactor"
)

var Ctx = context.Background()

var persistenceSet = wire.NewSet(
	logger.NewZapLogger,
	wire.Bind(new(domain.BlobStorage), new(*infrastructure.BlobS3Storage)),
	infrastructure.NewBlobS3Storage,
	provideContext,
	config.NewKernel,
	wire.Bind(new(domain.BlobRepository), new(*infrastructure.BlobDynamoRepository)),
	infrastructure.NewBlobDynamoRepository,
)

func provideContext() context.Context {
	return Ctx
}

func InjectBlobUseCase() (*interactor.Blob, func(), error) {
	wire.Build(
		persistenceSet,
		wire.Bind(new(domain.BlobEvent), new(*infrastructure.BlobKafkaEvent)),
		infrastructure.NewBlobKafkaEvent,
		interactor.NewBlob,
	)
	return &interactor.Blob{}, nil, nil
}

func InjectBlobSagaUseCase() (*interactor.BlobSAGA, func(), error) {
	wire.Build(
		persistenceSet,
		interactor.NewBlobSaga,
	)

	return &interactor.BlobSAGA{}, nil, nil
}
