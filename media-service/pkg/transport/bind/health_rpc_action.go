package bind

import (
	"context"
	"github.com/maestre3d/alexandria/media-service/pb"
	"google.golang.org/grpc"
)

type HealthRPCServer struct {
	srv pb.HealthServer
}

// Compile-time RPC implementation
type healthRPCImp struct{}

func NewHealthRPC() *HealthRPCServer {
	return &HealthRPCServer{healthRPCImp{}}
}

func (a HealthRPCServer) SetRoutes(srv *grpc.Server) {
	pb.RegisterHealthServer(srv, a.srv)
}

func (a healthRPCImp) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_SERVING}, nil
}
