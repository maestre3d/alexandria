package main

import (
	"fmt"
	"github.com/maestre3d/alexandria/media-service/pkg/shared/di"
	"github.com/oklog/run"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	transportService, cleanup, err := di.InjectTransportService()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// Run State Machines, API, Cron Job, Signal Handler
	// run.Group manages our goroutine lifecycles
	var g run.Group
	{
		l, err := net.Listen("tcp", transportService.HTTPProxy.Server.Addr)
		if err != nil {
			log.Fatalf("failed to start http server\nerror: %v", err)
		}

		g.Add(func() error {
			return http.Serve(l, transportService.HTTPProxy.Server.Handler)
		}, func(err error) {
			l.Close()
		})
	}
	{
		// The gRPC listener mounts the Go kit gRPC server we created.
		grpcListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", transportService.HTTPProxy.Config.TransportConfig.RPCHost,
			transportService.HTTPProxy.Config.TransportConfig.RPCPort))
		if err != nil {
			log.Fatalf("failed to start http server\nerror: %v", err)
		}
		g.Add(func() error {
			// we add the Go Kit gRPC Interceptor to our gRPC service as it is used by
			// the here demonstrated zipkin tracing middleware.
			return transportService.RPCProxy.Serve(grpcListener)
		}, func(error) {
			grpcListener.Close()
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
