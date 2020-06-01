package proxy

import (
	"fmt"
	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/maestre3d/alexandria/author-service/internal/infrastructure/config"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/pb"
	"google.golang.org/grpc"
)

type RPCProxyHandlers struct {
	AuthorHandler pb.AuthorServer
}

func NewRPCTransportProxy(proxyHandler *RPCProxyHandlers, logger log.Logger, cfg *config.KernelConfig) (*grpc.Server, func()) {
	// RPC Service registry
	rpcServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	pb.RegisterAuthorServer(rpcServer, proxyHandler.AuthorHandler)
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
