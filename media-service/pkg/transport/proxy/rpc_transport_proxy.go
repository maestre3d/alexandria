package proxy

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/media-service/internal/shared/infrastructure/config"
	"github.com/maestre3d/alexandria/media-service/pkg/transport/pb"
	"google.golang.org/grpc"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
)

type RPCProxyHandlers struct {
	MediaHandler pb.MediaServer
}

func NewRPCTransportProxy(proxyHandler *RPCProxyHandlers, logger log.Logger, cfg *config.KernelConfig) (*grpc.Server, func()) {
	// RPC Service registry
	rpcServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	pb.RegisterMediaServer(rpcServer, proxyHandler.MediaHandler)
	defer func() {
		logger.Log("method", "public.kernel.infrastructure.transport", "msg",
			fmt.Sprintf("grpc server created on addr %s:%d", cfg.TransportConfig.RPCHost,
				cfg.TransportConfig.RPCPort))
	}()

	cleanup := func() {
		rpcServer.Stop()
	}

	return rpcServer, cleanup
}
