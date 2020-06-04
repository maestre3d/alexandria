package service

import (
	"github.com/maestre3d/alexandria/author-service/pkg/transport/proxy"
	"google.golang.org/grpc"
)

type Transport struct {
	RPCProxy  *grpc.Server
	HTTPProxy *proxy.HTTP
}

func NewTransport(rpcProxy *grpc.Server, httpProxy *proxy.HTTP) *Transport {
	return &Transport{rpcProxy, httpProxy}
}
