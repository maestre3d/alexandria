package instrument

import (
	"context"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/pb/healthpb"
)

type HealthRPCServer struct{}

func NewHealthRPC() healthpb.HealthServer {
	return HealthRPCServer{}
}

func (a HealthRPCServer) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}
