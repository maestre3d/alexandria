package service

import (
	"github.com/alexandria-oss/core/config"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/proxy"
	"google.golang.org/grpc"
)

type Transport struct {
	RPCProxy  *grpc.Server
	HTTPProxy *proxy.HTTP
	Config    *config.Kernel
}

func NewTransport(rpcProxy *grpc.Server, httpProxy *proxy.HTTP, cfg *config.Kernel) *Transport {
	return &Transport{rpcProxy, httpProxy, cfg}
}
