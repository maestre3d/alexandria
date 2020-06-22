package service

import (
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/transport/proxy"
)

// Custom transport service

type Transport struct {
	EventProxy *proxy.Event
	Config     *config.Kernel
}

func NewTransport(eventProxy *proxy.Event, cfg *config.Kernel) *Transport {
	return &Transport{eventProxy, cfg}
}
