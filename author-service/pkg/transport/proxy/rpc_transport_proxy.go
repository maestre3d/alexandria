package proxy

import (
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/pb"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/pb/healthpb"
	"google.golang.org/grpc"
)

type Servers struct {
	AuthorServer pb.AuthorServer
	HealthServer healthpb.HealthServer
}

func NewRPC(servers *Servers) (*grpc.Server, func()) {
	// RPC Service registry
	rpcServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	pb.RegisterAuthorServer(rpcServer, servers.AuthorServer)
	healthpb.RegisterHealthServer(rpcServer, servers.HealthServer)
	cleanup := func() {
		rpcServer.Stop()
	}

	return rpcServer, cleanup
}
