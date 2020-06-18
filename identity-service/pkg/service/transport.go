package service

import (
	"github.com/alexandria-oss/core/config"
	"github.com/maestre3d/alexandria/identity-service/pkg/transport/proxy"
)

type Transport struct {
	EventProxy *proxy.Event
	Config     *config.Kernel
}

func NewTransport(eventProxy *proxy.Event, cfg *config.Kernel) *Transport {
	return &Transport{eventProxy, cfg}
}
