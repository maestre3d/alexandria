package transport

import (
	"github.com/maestre3d/alexandria/author-service/pkg/transport/proxy"
	"google.golang.org/grpc"
)

type TransportService struct {
	RPCProxy  *grpc.Server
	HTTPProxy *proxy.HTTPTransportProxy
}

func NewTransportService(rpcProxy *grpc.Server, httpProxy *proxy.HTTPTransportProxy) *TransportService {
	return &TransportService{rpcProxy, httpProxy}
}
