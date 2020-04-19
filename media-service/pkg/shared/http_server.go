package shared

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/media-service/internal/shared/infrastructure/config"
	"net/http"
	"time"
)

func NewHTTPServer(logger log.Logger, cfg *config.KernelConfig) *http.Server {
	r := mux.NewRouter()
	defer func() {
		logger.Log("method", "public.kernel.infrastructure.transport", "msg",
			fmt.Sprintf("http server created on addr %s:%d", cfg.TransportConfig.HTTPHost,
				cfg.TransportConfig.HTTPPort))
	}()

	return &http.Server{
		Addr:              cfg.TransportConfig.HTTPHost + fmt.Sprintf(":%d", cfg.TransportConfig.HTTPPort),
		Handler:           r,
		TLSConfig:         nil,
		ReadTimeout:       time.Second * 15,
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      time.Second * 15,
		IdleTimeout:       time.Second * 30,
		MaxHeaderBytes:    4096,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
		BaseContext:       nil,
		ConnContext:       nil,
	}
}
