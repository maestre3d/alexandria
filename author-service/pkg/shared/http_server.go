package shared

import (
	"fmt"
	"github.com/alexandria-oss/core/config"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func NewHTTPServer(cfg *config.Kernel) *http.Server {
	r := mux.NewRouter()

	return &http.Server{
		Addr:              cfg.Transport.HTTPHost + fmt.Sprintf(":%d", cfg.Transport.HTTPPort),
		Handler:           r,
		TLSConfig:         nil,
		ReadTimeout:       time.Second * 10,
		ReadHeaderTimeout: time.Second * 10,
		WriteTimeout:      time.Second * 10,
		IdleTimeout:       time.Second * 15,
		MaxHeaderBytes:    4096,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
		BaseContext:       nil,
		ConnContext:       nil,
	}
}
