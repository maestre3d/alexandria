package shared

import (
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func NewHTTPServer(logger log.Logger) *http.Server {
	r := mux.NewRouter()
	defer func() {
		logger.Log("method", "pkg.kernel.infrastructure.transport", "msg", "http server created")
	}()

	return &http.Server{
		Addr:              "0.0.0.0:8080",
		Handler:           r,
		TLSConfig:         nil,
		ReadTimeout:       time.Second * 15,
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      time.Second * 15,
		IdleTimeout:       time.Second * 30,
		MaxHeaderBytes:    8192,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
		BaseContext:       nil,
		ConnContext:       nil,
	}
}
