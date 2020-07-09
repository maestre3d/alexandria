package bind

import (
	"context"
	"encoding/json"
	"github.com/alexandria-oss/core"
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
	"github.com/maestre3d/alexandria/media-service/pkg/media/action"
	"github.com/maestre3d/alexandria/media-service/pkg/media/usecase"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strings"
)

type MediaHandler struct {
	service      usecase.MediaInteractor
	logger       log.Logger
	duration     *kitprometheus.Summary
	tracer       stdopentracing.Tracer
	zipkinTracer *stdzipkin.Tracer
	options      []httptransport.ServerOption
}

func NewMediaHTTP(svc usecase.MediaInteractor, logger log.Logger, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) *MediaHandler {
	duration := kitprometheus.NewSummaryFrom(prometheus.SummaryOpts{
		Namespace:   "alexandria",
		Subsystem:   "media_service",
		Name:        "request_duration_seconds",
		Help:        "total duration of requests in microseconds",
		ConstLabels: nil,
		Objectives:  nil,
		MaxAge:      0,
		AgeBuckets:  0,
		BufCap:      0,
	}, []string{"method", "success"})

	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(httputil.ResponseErrJSON),
		kitoc.HTTPServerTrace(),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	if zipkinTracer != nil {
		// Zipkin HTTP Server Trace can either be instantiated per endpoint with a
		// provided operation name or a global tracing usecase can be instantiated
		// without an operation name and fed to each Go kit endpoint as ServerOption.
		// In the latter case, the operation name will be the endpoint's http method.
		// We demonstrate a global tracing usecase here.
		options = append(options, zipkin.HTTPServerTrace(zipkinTracer, zipkin.Logger(logger), zipkin.Name("media_service"),
			zipkin.AllowPropagation(true)))
	}

	return &MediaHandler{svc, logger, duration, tracer, zipkinTracer, options}
}

// SetRoutes implement Handler interface for HTTP Proxy
func (h *MediaHandler) SetRoutes(public, private, admin *mux.Router) {
	// Admin routing
	arouter := admin.PathPrefix("/media").Subrouter()
	arouter.Methods(http.MethodOptions)
	arouter.Path("/{id}").Methods(http.MethodDelete).Handler(h.HardDelete())
	arouter.Use(mux.CORSMethodMiddleware(arouter))

	// Private routing
	pRouter := private.PathPrefix("/media").Subrouter()
	pRouter.Methods(http.MethodOptions)
	pRouter.Path("").Methods(http.MethodPost).Handler(h.Create())
	pRouter.Path("/").Methods(http.MethodPost).Handler(h.Create())

	pRouter.Path("/{id}").Methods(http.MethodPatch, http.MethodPut).Handler(h.Update())
	pRouter.Path("/{id}").Methods(http.MethodDelete).Handler(h.Delete())
	pRouter.Path("/{id}/restore").Methods(http.MethodPatch).Handler(h.Restore())
	pRouter.Use(mux.CORSMethodMiddleware(arouter))

	// Public routing
	r := public.PathPrefix("/media").Subrouter()
	r.Methods(http.MethodOptions)
	r.Path("").Methods(http.MethodGet).Handler(h.List())
	r.Path("/").Methods(http.MethodGet).Handler(h.List())

	r.Path("/{id}").Methods(http.MethodGet).Handler(h.Get())
	r.Use(mux.CORSMethodMiddleware(r))
}

func (h *MediaHandler) Create() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeCreateMediaEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeCreateRequest,
		encodeCreateResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Create", h.logger)))...,
	)
}

func (h *MediaHandler) List() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeListMediaEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeListRequest,
		encodeListResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "List", h.logger)))...,
	)
}

func (h *MediaHandler) Get() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeGetMediaEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeGetRequest,
		encodeGetResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Get", h.logger)))...,
	)
}

func (h *MediaHandler) Update() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeUpdateMediaEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeUpdateRequest,
		encodeUpdateResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Update", h.logger)))...,
	)
}

func (h *MediaHandler) Delete() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeDeleteMediaEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeDeleteRequest,
		encodeDeleteResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Delete", h.logger)))...,
	)
}

func (h *MediaHandler) Restore() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeRestoreMediaEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeRestoreRequest,
		encodeRestoreResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Restore", h.logger)))...,
	)
}

func (h *MediaHandler) HardDelete() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeHardDeleteMediaEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeHardDeleteRequest,
		encodeHardDeleteResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Hard_Delete", h.logger)))...,
	)
}

/* Decode HTTP Request */

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.CreateRequest{
		Title:        r.FormValue("title"),
		DisplayName:  r.FormValue("display_name"),
		Description:  r.FormValue("description"),
		LanguageCode: r.FormValue("language_code"),
		PublisherID:  r.FormValue("publisher_id"),
		AuthorID:     r.FormValue("author_id"),
		PublishDate:  r.FormValue("publish_date"),
		MediaType:    r.FormValue("media_type"),
	}, nil
}

func decodeListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.ListRequest{
		PageToken: r.URL.Query().Get("page_token"),
		PageSize:  r.URL.Query().Get("page_size"),
		FilterParams: core.FilterParams{
			"query":         r.URL.Query().Get("query"),
			"filter_by":     r.URL.Query().Get("filter_by"),
			"sort":          r.URL.Query().Get("sort"),
			"show_disabled": r.URL.Query().Get("show_disabled"),
			"lang":          r.URL.Query().Get("lang"),
			"publisher":     r.URL.Query().Get("publisher"),
			"author":        r.URL.Query().Get("author"),
			"media_type":    r.URL.Query().Get("media_type"),
		},
	}, nil
}

func decodeGetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.GetRequest{ID: mux.Vars(r)["id"]}, nil
}

func decodeUpdateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if strings.Contains(r.Header.Get("Content-Type"), "json") {
		var bodyJSON action.UpdateRequest
		err := json.NewDecoder(r.Body).Decode(&bodyJSON)
		if err == nil {
			return bodyJSON, nil
		}
	}

	return action.UpdateRequest{
		ID:           mux.Vars(r)["id"],
		Title:        r.FormValue("title"),
		DisplayName:  r.FormValue("display_name"),
		Description:  r.FormValue("description"),
		LanguageCode: r.FormValue("language_code"),
		PublisherID:  r.FormValue("publisher_id"),
		AuthorID:     r.FormValue("author_id"),
		PublishDate:  r.FormValue("publish_date"),
		MediaType:    r.FormValue("media_type"),
		URL:          r.FormValue("url"),
	}, nil
}

func decodeDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.DeleteRequest{ID: mux.Vars(r)["id"]}, nil
}

func decodeRestoreRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.RestoreRequest{ID: mux.Vars(r)["id"]}, nil
}

func decodeHardDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.HardDeleteRequest{ID: mux.Vars(r)["id"]}, nil
}

/* Encode HTTP Response */

func encodeCreateResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r := response.(action.CreateResponse)
	if r.Err != nil {
		httputil.ResponseErrJSON(ctx, r.Err, w)
		return nil
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeListResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.ListResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(ctx, r.Err, w)
			return nil
		} else if r.Err == nil && len(r.Medias) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return json.NewEncoder(w).Encode(httputil.GenericResponse{
				Message: exception.EntitiesNotFound.Error(),
				Code:    http.StatusNotFound,
			})
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeGetResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.GetResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(ctx, r.Err, w)
			return nil
		} else if r.Err == nil && r.Media == nil {
			w.WriteHeader(http.StatusNotFound)
			return json.NewEncoder(w).Encode(httputil.GenericResponse{
				Message: exception.EntityNotFound.Error(),
				Code:    http.StatusNotFound,
			})
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeUpdateResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.UpdateResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(ctx, r.Err, w)
			return nil
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeDeleteResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.DeleteResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(ctx, r.Err, w)
			return nil
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeRestoreResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.RestoreResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(ctx, r.Err, w)
			return nil
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeHardDeleteResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.HardDeleteResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(ctx, r.Err, w)
			return nil
		}
	}

	return json.NewEncoder(w).Encode(r)
}
