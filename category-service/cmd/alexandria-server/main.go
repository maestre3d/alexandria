package main

import (
	"context"
	"fmt"
	"github.com/maestre3d/alexandria/category-service/pkg/dep"
	"github.com/oklog/run"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())
	dep.SetContext(ctx)

	proxy, cleanup, err := dep.InjectTransportProxy()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		log.Print("stopping services")
		cleanup()
	}()

	var g run.Group
	{
		l, err := net.Listen("tcp", proxy.HTTP.Server.Addr)
		if err != nil {
			log.Fatal(err)
		}

		g.Add(func() error {
			log.Print("starting http server")
			return http.Serve(l, proxy.HTTP.Server.Handler)
		}, func(err error) {
			if err != nil {
				log.Print(err)
			}
			log.Print(l.Close())
		})
	}
	{
		// Set up signal bind
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
			// Cancel root context, propagate cancellation
			// cancel()
			close(cancelInterrupt)
		})
	}

	log.Fatal(g.Run())
}
