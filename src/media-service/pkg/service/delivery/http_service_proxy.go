package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/src/media-service/pkg/service/delivery/handler"
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
	MediaHandler *handler.MediaHandler
}

func NewHTTPServiceProxy(logger util.ILogger, server *http.Server, handlers *ProxyHandlers) *HTTPServiceProxy {
	var engine *gin.Engine
	engine, ok := server.Handler.(*gin.Engine)
	if !ok {
		// Replace handler with current (gin)
		server.Handler = gin.Default()
		engine = server.Handler.(*gin.Engine)
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

	logger.Print("http proxy service started", "service.delivery")

	return service
}

func (p *HTTPServiceProxy) mapBookRoutes() {
	bookRouter := p.publicGroup.Group("/media")

	bookRouter.POST("", p.proxyHandlers.MediaHandler.Create)
	bookRouter.GET("", p.proxyHandlers.MediaHandler.List)
	bookRouter.GET("/:media_id", p.proxyHandlers.MediaHandler.Get)
	bookRouter.PATCH("/:media_id", p.proxyHandlers.MediaHandler.UpdateOne)
	bookRouter.DELETE("/:media_id", p.proxyHandlers.MediaHandler.DeleteOne)
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
