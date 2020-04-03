package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func main() {
	engine := gin.Default()

	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, &gin.H{
			"message": "Hello from Book HTTP API",
		})
	})

	server := http.Server{
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

	err := server.ListenAndServe()
	defer func() {
		err := server.Shutdown(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	if err != nil {
		panic(err)
	}
}
