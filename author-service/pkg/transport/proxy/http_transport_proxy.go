package proxy

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/httputil"
	"github.com/rs/cors"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler interface {
	SetRoutes(public, private, admin *mux.Router)
}

type HTTP struct {
	Server        *http.Server
	publicRouter  *mux.Router
	privateRouter *mux.Router
	adminRouter   *mux.Router
	handlers      []Handler
}

func NewHTTP(cfg *config.Kernel, handlers ...Handler) (*HTTP, func()) {
	r := mux.NewRouter()
	server := httputil.DefaultServer(cfg, r)

	proxy := &HTTP{
		Server:        server,
		publicRouter:  newHTTPPublicRouter(r),
		privateRouter: newHTTPPrivateRouter(r),
		adminRouter:   newHTTPAdminRouter(r),
		handlers:      handlers,
	}

	proxy.setHealthCheck()
	proxy.setMetrics()

	proxy.mapRoutes()

	proxy.Server.Handler = cors.Default().Handler(r)

	cleanup := func() {
		_ = server.Shutdown(context.Background())
	}

	return proxy, cleanup
}

func (p *HTTP) setHealthCheck() {
	p.publicRouter.PathPrefix("/healthpb").Methods(http.MethodGet).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(writer, `{"alive":true}`)
	})
}

func (p *HTTP) setMetrics() {
	p.adminRouter.PathPrefix("/metrics").Methods(http.MethodGet).Handler(promhttp.Handler())
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
