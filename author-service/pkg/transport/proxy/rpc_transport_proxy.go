package proxy

import (
	"fmt"
	"github.com/alexandria-oss/core/config"
	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/pb"
	"google.golang.org/grpc"
)

type RPC struct {
	AuthorServer pb.AuthorServer
}

func NewRPC(proxyHandler *RPC, logger log.Logger, cfg *config.Kernel) (*grpc.Server, func()) {
	// RPC Service registry
	rpcServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	pb.RegisterAuthorServer(rpcServer, proxyHandler.AuthorServer)
	defer func() {
		logger.Log("method", "public.kernel.infrastructure.transport", "msg",
			fmt.Sprintf("grpc server created on addr %s:%d", cfg.Transport.RPCHost,
				cfg.Transport.RPCPort))
	}()

	cleanup := func() {
		rpcServer.Stop()
	}

	return rpcServer, cleanup
}
