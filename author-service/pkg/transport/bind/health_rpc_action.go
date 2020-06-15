package bind

import (
	"context"
	"github.com/maestre3d/alexandria/author-service/pb"
)

type HealthRPCServer struct{}

func NewHealthRPC() pb.HealthServer {
	return HealthRPCServer{}
}

func (a HealthRPCServer) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{Status: pb.HealthCheckResponse_SERVING}, nil
}
