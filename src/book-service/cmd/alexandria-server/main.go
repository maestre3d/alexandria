package main

import (
	"fmt"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/infrastructure/dependency"
	"github.com/oklog/run"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	httpService, clean, err := dependency.InitHTTPServiceProxy()
	if err != nil {
		panic(err)
	}
	defer clean()

	// Run State Machines, API, Cron Job, Signal Handler
	// run.Group manages our goroutine lifecycles
	var g run.Group
	{
		var listener, _ = net.Listen("tcp", ":8080") // might use port :0 for dynamic port assignment
		g.Add(func() error {
			return http.Serve(listener, httpService.Server.Handler)
		}, func(err error) {
			listener.Close()
		})
	}
	{
		// Set-up signal handler
		var (
			cancelInterrupt = make(chan struct{})
			c               = make(chan os.Signal, 2)
		)
		defer close(c)

		g.Add(func() error {
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}

	g.Run()
}
