package proxy

import (
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

type RPCServer interface {
	SetRoutes(*grpc.Server)
}

func NewRPC(servers []RPCServer) (*grpc.Server, func()) {
	// RPC Service registry
	rpcServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	mapRoutes(servers, rpcServer)

	cleanup := func() {
		rpcServer.Stop()
	}

	return rpcServer, cleanup
}

func mapRoutes(servers []RPCServer, rpc *grpc.Server) {
	for _, srv := range servers {
		srv.SetRoutes(rpc)
	}
}
