package observability

import (
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"net/http"
)

func Trace(h http.HandlerFunc, isPublic bool) http.Handler {
	return &ochttp.Handler{
		Propagation:      nil,
		Handler:          h,
		StartOptions:     trace.StartOptions{},
		GetStartOptions:  nil,
		IsPublicEndpoint: isPublic,
		FormatSpanName:   nil,
		IsHealthEndpoint: nil,
	}
}
