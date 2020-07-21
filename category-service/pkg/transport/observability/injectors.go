package observability

import (
	ocprom "contrib.go.opencensus.io/exporter/prometheus"
	oczipkin "contrib.go.opencensus.io/exporter/zipkin"
	"github.com/alexandria-oss/core/config"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	"github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
)

func InjectPrometheus(cfg *config.Kernel) (*ocprom.Exporter, error) {
	err := view.Register(
		ochttp.ServerLatencyView,
		ochttp.ServerRequestBytesView,
		ochttp.ServerRequestCountByMethod,
		ochttp.ServerRequestCountView,
		ochttp.ServerResponseBytesView,
		ochttp.ServerResponseCountByStatusCode,
	)
	if err := view.Register(ochttp.ServerLatencyView); err != nil {
		return nil, err
	}

	pe, err := ocprom.NewExporter(ocprom.Options{
		Namespace:   cfg.Service,
		Registry:    nil,
		Registerer:  nil,
		Gatherer:    nil,
		OnError:     nil,
		ConstLabels: nil,
	})
	if err != nil {
		return nil, err
	}

	view.RegisterExporter(pe)

	return pe, nil
}

func InjectZipkin(cfg *config.Kernel) error {
	localEndpoint, err := zipkin.NewEndpoint(cfg.Service, cfg.Tracing.ZipkinEndpoint)
	if err != nil {
		return err
	}
	reporter := zipkinHTTP.NewReporter(cfg.Tracing.ZipkinHost)
	ze := oczipkin.NewExporter(reporter, localEndpoint)
	trace.RegisterExporter(ze)
	trace.ApplyConfig(trace.Config{
		DefaultSampler:             trace.AlwaysSample(),
		IDGenerator:                nil,
		MaxAnnotationEventsPerSpan: 0,
		MaxMessageEventsPerSpan:    0,
		MaxAttributesPerSpan:       0,
		MaxLinksPerSpan:            0,
	})
	return nil
}
