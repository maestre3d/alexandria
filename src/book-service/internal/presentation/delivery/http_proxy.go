package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/maestre3d/alexandria/src/book-service/internal/book/application"
	"github.com/maestre3d/alexandria/src/book-service/internal/book/domain"
	"github.com/maestre3d/alexandria/src/book-service/internal/book/infrastructure"
	"github.com/maestre3d/alexandria/src/book-service/internal/presentation/delivery/handler"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/util"
)

// InitHTTPPublicProxy Start HTTP Service's public proxy
func InitHTTPPublicProxy(logger util.ILogger, engine *gin.Engine) error {
	publicGroup := engine.Group(global.PublicAPI)

	// Init required Handlers
	// TODO: Use Google's wire DI to correctly inject dependencies
	// TODO: Add use cases
	bookLocalRepo := infrastructure.NewBookLocalRepository(make([]*domain.BookEntity, 0))

	err := handler.NewBookHandler(logger, publicGroup, application.NewBookUseCase(logger, bookLocalRepo))
	if err != nil {
		return err
	}

	return nil
}
