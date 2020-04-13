package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/media-service/internal/shared/infrastructure/config"
	"net/http"
	"time"
)

func NewHTTPServer(logger util.ILogger, cfg *config.KernelConfig) *http.Server {
	gin.SetMode("release")
	engine := gin.Default()
	logger.Print("http server created", "service.delivery")

	return &http.Server{
		Addr:              cfg.HTTPPort,
		Handler:           engine,
		TLSConfig:         nil,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    2048,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
		BaseContext:       nil,
		ConnContext:       nil,
	}
}
