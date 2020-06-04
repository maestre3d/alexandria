package proxy

import (
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/pb"
	"google.golang.org/grpc"
)

type Servers struct {
	AuthorServer pb.AuthorServer
}

func NewRPC(servers *Servers) (*grpc.Server, func()) {
	// RPC Service registry
	rpcServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	pb.RegisterAuthorServer(rpcServer, servers.AuthorServer)
	cleanup := func() {
		rpcServer.Stop()
	}

	return rpcServer, cleanup
}
