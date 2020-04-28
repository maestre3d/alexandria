package proxy

import (
	"context"
	"io"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/media-service/internal/shared/infrastructure/config"
	"github.com/maestre3d/alexandria/media-service/pkg/transport/handler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPTransportProxy struct {
	Server        *http.Server
	Config        *config.KernelConfig
	publicRouter  *mux.Router
	privateRouter *mux.Router
	adminRouter   *mux.Router
	logger        log.Logger
	handlers      *ProxyHandlers
}

type ProxyHandlers struct {
	MediaHandler *handler.MediaHandler
}

func NewHTTPTransportProxy(logger log.Logger, server *http.Server, cfg *config.KernelConfig, handlers *ProxyHandlers) (*HTTPTransportProxy, func()) {
	// TODO: Add metrics with OpenCensus and Prometheus/Zipkin
	router, ok := server.Handler.(*mux.Router)
	if !ok {
		server.Handler = mux.NewRouter()
		router = server.Handler.(*mux.Router)
	}

	proxy := &HTTPTransportProxy{
		Server:        server,
		Config:        cfg,
		publicRouter:  newHTTPPublicRouter(router),
		privateRouter: newHTTPPrivateRouter(router),
		adminRouter:   newHTTPAdminRouter(router),
		logger:        logger,
		handlers:      handlers,
	}

	// TODO: Change public policies to admin
	proxy.setHealthCheck()
	proxy.setMetrics()

	proxy.mapRoutes()

	cleanup := func() {
		server.Shutdown(context.Background())
	}

	return proxy, cleanup
}

func (p *HTTPTransportProxy) setHealthCheck() {
	p.publicRouter.PathPrefix("/health").Methods(http.MethodGet).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		io.WriteString(writer, `{"alive":true}`)
	})
}

func (p *HTTPTransportProxy) setMetrics() {
	p.publicRouter.PathPrefix("/metrics").Methods(http.MethodGet).Handler(promhttp.Handler())
}

func (p *HTTPTransportProxy) mapRoutes() {
	authorRouter := p.publicRouter.PathPrefix("/media").Subrouter()
	authorRouter.Path("").Methods(http.MethodPost).Handler(p.handlers.MediaHandler.Create())
	authorRouter.Path("").Methods(http.MethodGet).Handler(p.handlers.MediaHandler.List())
	authorRouter.Path("/").Methods(http.MethodPost).Handler(p.handlers.MediaHandler.Create())
	authorRouter.Path("/").Methods(http.MethodGet).Handler(p.handlers.MediaHandler.List())

	authorRouter.Path("/{id}").Methods(http.MethodGet).Handler(p.handlers.MediaHandler.Get())
	authorRouter.Path("/{id}").Methods(http.MethodPatch, http.MethodPut).Handler(p.handlers.MediaHandler.Update())
	authorRouter.Path("/{id}").Methods(http.MethodDelete).Handler(p.handlers.MediaHandler.Delete())
}

func newHTTPPublicRouter(r *mux.Router) *mux.Router {
	return r.PathPrefix(global.PublicAPI).Subrouter()
}

func newHTTPPrivateRouter(r *mux.Router) *mux.Router {
	return r.PathPrefix(global.PrivateAPI).Subrouter()
}

func newHTTPAdminRouter(r *mux.Router) *mux.Router {
	return r.PathPrefix(global.AdminAPI).Subrouter()
}
