package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/src/book-service/pkg/service/delivery/handler"
	"net/http"
)

type HTTPServiceProxy struct {
	Server        *http.Server
	publicGroup   *gin.RouterGroup
	privateGroup  *gin.RouterGroup
	adminGroup    *gin.RouterGroup
	proxyHandlers *ProxyHandlers
	logger        util.ILogger
}

type ProxyHandlers struct {
	BookHandler *handler.BookHandler
}

func NewHTTPServiceProxy(logger util.ILogger, server *http.Server, handlers *ProxyHandlers) *HTTPServiceProxy {
	engine, ok := server.Handler.(*gin.Engine)
	// TODO: Throw err
	if !ok {
		return nil
	}
	service := &HTTPServiceProxy{
		Server:        server,
		publicGroup:   newHTTPPublicProxy(logger, engine),
		privateGroup:  newHTTPPrivateProxy(logger, engine),
		adminGroup:    newHTTPAdminProxy(logger, engine),
		proxyHandlers: handlers,
		logger:        logger,
	}

	// Start routing-mapping
	service.mapBookRoutes()

	return service
}

func (p *HTTPServiceProxy) mapBookRoutes() {
	bookRouter := p.publicGroup.Group("/book")

	bookRouter.POST("/", p.proxyHandlers.BookHandler.Create)
	bookRouter.GET("/", p.proxyHandlers.BookHandler.List)
	bookRouter.GET("/:book_id", p.proxyHandlers.BookHandler.Get)
	bookRouter.PATCH("/:book_id", p.proxyHandlers.BookHandler.UpdateOne)
	bookRouter.DELETE("/:book_id", p.proxyHandlers.BookHandler.DeleteOne)
}

// InitHTTPPublicProxy Start HTTP Service's public proxy
func newHTTPPublicProxy(logger util.ILogger, engine *gin.Engine) *gin.RouterGroup {
	publicGroup := engine.Group(global.PublicAPI)

	return publicGroup
}

func newHTTPPrivateProxy(logger util.ILogger, engine *gin.Engine) *gin.RouterGroup {
	publicGroup := engine.Group(global.PrivateAPI)

	return publicGroup
}

func newHTTPAdminProxy(logger util.ILogger, engine *gin.Engine) *gin.RouterGroup {
	publicGroup := engine.Group(global.AdminAPI)

	return publicGroup
}
