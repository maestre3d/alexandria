package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/util"
	"net/http"
	"time"
)

func NewHTTPServer(logger util.ILogger) *http.Server {
	gin.SetMode("release")
	engine := gin.Default()
	logger.Print("http server created", "service.delivery")

	return &http.Server{
		Addr:              ":8080",
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
