package transport

import (
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/config"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	muxhandler "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/category-service/pkg/transport/observability"
	"net/http"
	"os"
	"time"
)

type Handler interface {
	SetRoutes(public, private, admin *mux.Router)
	GetName() string
}

type HTTPServer struct {
	Server    *http.Server
	Cfg       *config.Kernel
	logger    log.Logger
	handlers  []Handler
	router    *mux.Router
	startTime time.Time
}

func NewHTTPServer(cfg *config.Kernel, logger log.Logger, handlers ...Handler) *HTTPServer {
	// Start and set router configs
	router := mux.NewRouter()
	router.Use(muxhandler.RecoveryHandler())
	router.Use(muxhandler.CORS(
		muxhandler.AllowedMethods([]string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		}),
		muxhandler.AllowedOrigins([]string{"*"}),
	))
	router.Use(muxhandler.CompressHandler)

	httpServer := &HTTPServer{handlers: handlers, Cfg: cfg, logger: logger, router: router, startTime: time.Now()}

	// Inject metrics w OpenCensus and Prometheus
	pe, err := observability.InjectPrometheus(cfg)
	if err != nil {
		_ = level.Error(httpServer.logger).Log(
			"msg", "cannot start prometheus http metrics",
			"err", err,
		)
		return nil
	}
	router.Path("/metrics").Handler(pe)

	// Inject distributed tracing w OpenCensus and Zipkin
	err = observability.InjectZipkin(cfg)
	if err != nil {
		_ = level.Error(httpServer.logger).Log(
			"msg", "cannot start OpenCensus and Zipkin exporter for distributed tracing",
			"err", err,
		)
		return nil
	}

	// Inject kubernetes liveness endpoint for health check
	router.Path("/healthz").HandlerFunc(httpServer.HealthCheck)

	// Start router-handler mapping
	httpServer.setRoutes()

	httpServer.Server = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Transport.HTTPHost, cfg.Transport.HTTPPort),
		Handler:           muxhandler.CombinedLoggingHandler(os.Stdout, router),
		TLSConfig:         nil,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 0,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
		BaseContext:       nil,
		ConnContext:       nil,
	}

	return httpServer
}

func (s *HTTPServer) setRoutes() {
	public := s.router.PathPrefix(core.PublicAPI).Subrouter()
	private := s.router.PathPrefix(core.PrivateAPI).Subrouter()
	admin := s.router.PathPrefix(core.AdminAPI).Subrouter()

	for _, handler := range s.handlers {
		handler.SetRoutes(public, private, admin)
		_ = level.Info(s.logger).Log(
			"msg", fmt.Sprintf("http handler %s mapped", handler.GetName()),
		)
	}

	_ = level.Info(s.logger).Log(
		"msg", "http-handler mapping has been successful",
	)
}

func (s HTTPServer) HealthCheck(w http.ResponseWriter, r *http.Request) {
	duration := time.Now().Sub(s.startTime)
	if duration.Seconds() > 10 {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(fmt.Sprintf("error: %v", duration.Seconds())))
	} else {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}
}
