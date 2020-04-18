package main

import (
	"fmt"
	"github.com/maestre3d/alexandria/author-service/pkg/shared/di"
	"github.com/oklog/run"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	proxyHTTP, cleanup, err := di.InjectHTTPProxy()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// Manage goroutines
	var g run.Group
	{
		l, _ := net.Listen("tcp", ":8080")
		g.Add(func() error {
			return http.Serve(l, proxyHTTP.Server.Handler)
		}, func(err error) {
			l.Close()
		})
	}
	{
		// Set up signal handler
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
