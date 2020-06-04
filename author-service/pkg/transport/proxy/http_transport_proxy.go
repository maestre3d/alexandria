package proxy

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/config"
	"io"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler interface {
	SetRoutes(public, private, admin *mux.Router)
}

type HTTP struct {
	Server        *http.Server
	Config        *config.Kernel
	publicRouter  *mux.Router
	privateRouter *mux.Router
	adminRouter   *mux.Router
	logger        log.Logger
	handlers      []Handler
}

func NewHTTPTransportProxy(logger log.Logger, server *http.Server, cfg *config.Kernel, handlers ...Handler) (*HTTP, func()) {
	router, ok := server.Handler.(*mux.Router)
	if !ok {
		server.Handler = mux.NewRouter()
		router = server.Handler.(*mux.Router)
	}

	proxy := &HTTP{
		Server:        server,
		Config:        cfg,
		publicRouter:  newHTTPPublicRouter(router),
		privateRouter: newHTTPPrivateRouter(router),
		adminRouter:   newHTTPAdminRouter(router),
		logger:        logger,
		handlers:      handlers,
	}

	proxy.setHealthCheck()
	// TODO: Change public policies to admin
	proxy.setMetrics()

	proxy.mapRoutes()

	cleanup := func() {
		server.Shutdown(context.Background())
	}

	return proxy, cleanup
}

func (p *HTTP) setHealthCheck() {
	p.publicRouter.PathPrefix("/health").Methods(http.MethodGet).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		io.WriteString(writer, `{"alive":true}`)
	})
}

func (p *HTTP) setMetrics() {
	p.publicRouter.PathPrefix("/metrics").Methods(http.MethodGet).Handler(promhttp.Handler())
}

func (p *HTTP) mapRoutes() {
	for _, handler := range p.handlers {
		handler.SetRoutes(p.publicRouter, p.privateRouter, p.adminRouter)
	}
}

func newHTTPPublicRouter(r *mux.Router) *mux.Router {
	return r.PathPrefix(core.PublicAPI).Subrouter()
}

func newHTTPPrivateRouter(r *mux.Router) *mux.Router {
	return r.PathPrefix(core.PrivateAPI).Subrouter()
}

func newHTTPAdminRouter(r *mux.Router) *mux.Router {
	return r.PathPrefix(core.AdminAPI).Subrouter()
}
