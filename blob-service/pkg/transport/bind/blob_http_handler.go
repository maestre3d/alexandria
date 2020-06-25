package bind

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/alexandria-oss/core/httputil"
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	kitoc "github.com/go-kit/kit/tracing/opencensus"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob/action"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob/usecase"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/prometheus/client_golang/prometheus"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type BlobHandler struct {
	svc          usecase.BlobInteractor
	logger       log.Logger
	duration     *kitprometheus.Summary
	tracer       stdopentracing.Tracer
	zipkinTracer *stdzipkin.Tracer
	options      []httptransport.ServerOption
}

func NewBlobHandler(svc usecase.BlobInteractor, logger log.Logger, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) *BlobHandler {
	duration := kitprometheus.NewSummaryFrom(prometheus.SummaryOpts{
		Namespace:   "alexandria",
		Subsystem:   "blob_service",
		Name:        "request_duration_seconds",
		Help:        "total duration of requests in microseconds",
		ConstLabels: nil,
		Objectives:  nil,
		MaxAge:      0,
		AgeBuckets:  0,
		BufCap:      0,
	}, []string{"method", "success"})

	// Add OpenCensus metrics
	// Add error encoder
	// Add error logger
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(httputil.ResponseErrJSON),
		kitoc.HTTPServerTrace(),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	// Inject tracing exporter
	if zipkinTracer != nil {
		options = append(options, zipkin.HTTPServerTrace(zipkinTracer))
	}

	return &BlobHandler{
		svc:          svc,
		logger:       logger,
		duration:     duration,
		tracer:       tracer,
		zipkinTracer: zipkinTracer,
		options:      options,
	}
}

func (h *BlobHandler) SetRoutes(public, private, _ *mux.Router) {
	// Private routing
	mediaP := private.PathPrefix("/blob/" + domain.Media).Subrouter()
	mediaP.Methods(http.MethodOptions)
	authorP := private.PathPrefix("/blob/" + domain.Author).Subrouter()
	authorP.Methods(http.MethodOptions)
	userP := private.PathPrefix("/blob/" + domain.User).Subrouter()
	userP.Methods(http.MethodOptions)

	mediaP.Path("/{id}").Methods(http.MethodPost).Handler(h.Store())
	mediaP.Path("/{id}").Methods(http.MethodDelete).Handler(h.Delete())
	mediaP.Use(mux.CORSMethodMiddleware(mediaP))

	authorP.Path("/{id}").Methods(http.MethodPost).Handler(h.Store())
	authorP.Path("/{id}").Methods(http.MethodDelete).Handler(h.Delete())
	authorP.Use(mux.CORSMethodMiddleware(authorP))

	userP.Path("/{id}").Methods(http.MethodPost).Handler(h.Store())
	userP.Path("/{id}").Methods(http.MethodDelete).Handler(h.Delete())
	userP.Use(mux.CORSMethodMiddleware(userP))

	// Public routing
	mediaR := public.PathPrefix("/blob/" + domain.Media).Subrouter()
	mediaR.Methods(http.MethodOptions)
	authorR := public.PathPrefix("/blob/" + domain.Author).Subrouter()
	authorR.Methods(http.MethodOptions)
	userR := public.PathPrefix("/blob/" + domain.User).Subrouter()
	userR.Methods(http.MethodOptions)

	mediaR.Path("/{id}").Methods(http.MethodGet).Handler(h.Get())
	mediaR.Use(mux.CORSMethodMiddleware(mediaR))

	authorR.Path("/{id}").Methods(http.MethodGet).Handler(h.Get())
	authorR.Use(mux.CORSMethodMiddleware(authorR))

	userR.Path("/{id}").Methods(http.MethodGet).Handler(h.Get())
	userR.Use(mux.CORSMethodMiddleware(userR))
}

func (h *BlobHandler) Store() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeStoreBlobEndpoint(h.svc, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeStoreRequest,
		encodeStoreResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Store", h.logger)))...,
	)
}

func (h *BlobHandler) Get() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeGetBlobEndpoint(h.svc, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeGetRequest,
		encodeGetResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Get", h.logger)))...,
	)
}

func (h *BlobHandler) Delete() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeDeleteBlobEndpoint(h.svc, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeDeleteRequest,
		encodeDeleteResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Delete", h.logger)))...,
	)
}

/* Decoders */

func decodeStoreRequest(_ context.Context, r *http.Request) (interface{}, error) {
	f, h, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	contentSlice := strings.Split(h.Header.Get("Content-Type"), "/")
	if len(contentSlice) <= 1 {
		return nil, exception.NewErrorDescription(exception.InvalidFieldFormat,
			fmt.Sprintf(exception.InvalidFieldFormatString,
				"file", "invalid file extension"))
	}

	// go's file.size is in bytes
	var maxSize int64
	maxSize = 8192 * (int64(math.Pow(1024, 2)))

	if h.Size > maxSize {
		return nil, exception.NewErrorDescription(exception.InvalidFieldRange,
			fmt.Sprintf(exception.InvalidFieldRangeString,
				"file", "1 B", "8 GB"))
	}

	return action.StoreRequest{
		RootID:    mux.Vars(r)["id"],
		Service:   getServiceFromPath(r.URL.Path),
		BlobType:  contentSlice[0],
		Extension: contentSlice[1],
		Size:      strconv.Itoa(int(h.Size)),
		Content:   f,
	}, nil
}

func decodeGetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.GetRequest{
		ID:      mux.Vars(r)["id"],
		Service: getServiceFromPath(r.URL.Path),
	}, nil
}

func decodeDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.DeleteRequest{
		ID:      mux.Vars(r)["id"],
		Service: getServiceFromPath(r.URL.Path),
	}, nil
}

/* Encoders */

func encodeStoreResponse(ctx context.Context, w http.ResponseWriter, res interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := res.(action.StoreResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(ctx, r.Err, w)
			return nil
		}
	}

	return json.NewEncoder(w).Encode(r.Blob)
}

func encodeGetResponse(ctx context.Context, w http.ResponseWriter, res interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := res.(action.GetResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(ctx, r.Err, w)
			return nil
		}
	}

	return json.NewEncoder(w).Encode(r.Blob)
}

func encodeDeleteResponse(ctx context.Context, w http.ResponseWriter, res interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := res.(action.DeleteResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(ctx, r.Err, w)
			return nil
		}
	}

	return json.NewEncoder(w).Encode(struct{}{})
}
