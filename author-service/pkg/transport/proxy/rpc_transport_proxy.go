package proxy

import (
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/pb"
	"google.golang.org/grpc"
)

type RPCProxyHandlers struct {
	AuthorHandler pb.AuthorServer
}

func NewRPCTransportProxy(proxyHandler *RPCProxyHandlers) (*grpc.Server, func()) {
	// RPC Service registry
	rpcServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	pb.RegisterAuthorServer(rpcServer, proxyHandler.AuthorHandler)

	cleanup := func() {
		rpcServer.Stop()
	}

	return rpcServer, cleanup
}
