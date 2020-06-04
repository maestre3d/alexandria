package bind

import (
	"context"
	"encoding/json"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"github.com/alexandria-oss/core/httputil"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	kitoc "github.com/go-kit/kit/tracing/opencensus"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"net/http"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/author-service/pkg/author/action"
)

type AuthorHandler struct {
	service      usecase.AuthorInteractor
	logger       log.Logger
	duration     *kitprometheus.Summary
	tracer       stdopentracing.Tracer
	zipkinTracer *stdzipkin.Tracer
	options      []httptransport.ServerOption
}

func NewAuthorHandler(svc usecase.AuthorInteractor, logger log.Logger, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) *AuthorHandler {
	duration := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace:   "alexandria",
		Subsystem:   "author_service",
		Name:        "request_duration_seconds",
		Help:        "total duration of requests in microseconds",
		ConstLabels: nil,
		Objectives:  nil,
		MaxAge:      0,
		AgeBuckets:  0,
		BufCap:      0,
	}, []string{"method", "success"})

	options := []httptransport.ServerOption{
		// httptransport.ServerErrorEncoder(helper.ErrorEncoder),
		kitoc.HTTPServerTrace(),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	if zipkinTracer != nil {
		// Zipkin HTTP Server Trace can either be instantiated per endpoint with a
		// provided operation name or a global tracing usecase can be instantiated
		// without an operation name and fed to each Go kit endpoint as ServerOption.
		// In the latter case, the operation name will be the endpoint's http method.
		// We demonstrate a global tracing usecase here.
		options = append(options, zipkin.HTTPServerTrace(zipkinTracer))
	}

	return &AuthorHandler{svc, logger, duration, tracer, zipkinTracer, options}
}

// SetRoutes implement Handler interface for HTTP Proxy
func (h *AuthorHandler) SetRoutes(public, private, admin *mux.Router) {
	// Admin routing
	arouter := admin.PathPrefix("/author").Subrouter()
	arouter.Path("/{id}").Methods(http.MethodDelete).Handler(h.HardDelete())
	arouter.Path("/{id}").Methods(http.MethodPatch).Handler(h.Restore())

	// Public routing
	r := public.PathPrefix("/author").Subrouter()
	r.Path("").Methods(http.MethodPost).Handler(h.Create())
	r.Path("").Methods(http.MethodGet).Handler(h.List())
	r.Path("/").Methods(http.MethodPost).Handler(h.Create())
	r.Path("/").Methods(http.MethodGet).Handler(h.List())

	r.Path("/{id}").Methods(http.MethodGet).Handler(h.Get())
	r.Path("/{id}").Methods(http.MethodPatch, http.MethodPut).Handler(h.Update())
	r.Path("/{id}").Methods(http.MethodDelete).Handler(h.Delete())
}

func (h *AuthorHandler) Create() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeCreateAuthorEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeCreateRequest,
		encodeCreateResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Create", h.logger)))...,
	)
}

func (h *AuthorHandler) List() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeListAuthorEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeListRequest,
		encodeListResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "List", h.logger)))...,
	)
}

func (h *AuthorHandler) Get() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeGetAuthorEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeGetRequest,
		encodeGetResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Get", h.logger)))...,
	)
}

func (h *AuthorHandler) Update() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeUpdateAuthorEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeUpdateRequest,
		encodeUpdateResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Update", h.logger)))...,
	)
}

func (h *AuthorHandler) Delete() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeDeleteAuthorEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeDeleteRequest,
		encodeDeleteResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Delete", h.logger)))...,
	)
}

func (h *AuthorHandler) Restore() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeRestoreAuthorEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeRestoreRequest,
		encodeRestoreResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "Restore", h.logger)))...,
	)
}

func (h *AuthorHandler) HardDelete() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeHardDeleteAuthorEndpoint(h.service, h.logger, h.duration, h.tracer, h.zipkinTracer),
		decodeHardDeleteRequest,
		encodeHardDeleteResponse,
		append(h.options, httptransport.ServerBefore(opentracing.HTTPToContext(h.tracer, "HardDelete", h.logger)))...,
	)
}

/* Decode HTTP Request */

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.CreateRequest{
		FirstName:     r.PostFormValue("first_name"),
		LastName:      r.PostFormValue("last_name"),
		DisplayName:   r.PostFormValue("display_name"),
		OwnershipType: r.PostFormValue("ownership_type"),
		OwnerID:       r.PostFormValue("owner_id"),
	}, nil
}

func decodeListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.ListRequest{
		PageToken: r.URL.Query().Get("page_token"),
		PageSize:  r.URL.Query().Get("page_size"),
		FilterParams: core.FilterParams{
			"query":     r.URL.Query().Get("query"),
			"timestamp": r.URL.Query().Get("timestamp"),
		},
	}, nil
}

func decodeGetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.GetRequest{mux.Vars(r)["id"]}, nil
}

func decodeUpdateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var bodyJSON action.UpdateRequest
	err := json.NewDecoder(r.Body).Decode(&bodyJSON)
	if err != nil {
		return action.UpdateRequest{
			ID:            mux.Vars(r)["id"],
			FirstName:     r.PostFormValue("first_name"),
			LastName:      r.PostFormValue("last_name"),
			DisplayName:   r.PostFormValue("display_name"),
			OwnershipType: r.PostFormValue("ownership_type"),
			OwnerID:       r.PostFormValue("owner_id"),
			Verified:      r.PostFormValue("verified"),
			Picture:       r.PostFormValue("picture"),
			Owners:        nil,
		}, nil
	}

	return bodyJSON, nil
}

func decodeDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.DeleteRequest{mux.Vars(r)["id"]}, nil
}

func decodeRestoreRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.RestoreRequest{mux.Vars(r)["id"]}, nil
}

func decodeHardDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.HardDeleteRequest{mux.Vars(r)["id"]}, nil
}

/* Encode HTTP Response */

func encodeCreateResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.CreateResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(w, r.Err)
			return nil
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeListResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.ListResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(w, r.Err)
			return nil
		} else if r.Err == nil && len(r.Authors) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return json.NewEncoder(w).Encode(httputil.GenericResponse{
				Message: exception.EntitiesNotFound.Error(),
				Code:    http.StatusNotFound,
			})
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeGetResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.GetResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(w, r.Err)
			return nil
		} else if r.Err == nil && r.Author == nil {
			w.WriteHeader(http.StatusNotFound)
			return json.NewEncoder(w).Encode(httputil.GenericResponse{
				Message: exception.EntityNotFound.Error(),
				Code:    http.StatusNotFound,
			})
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeUpdateResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.UpdateResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(w, r.Err)
			return nil
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeDeleteResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.DeleteResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(w, r.Err)
			return nil
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeRestoreResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.RestoreResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(w, r.Err)
			return nil
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeHardDeleteResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.HardDeleteResponse)
	if ok {
		if r.Err != nil {
			httputil.ResponseErrJSON(w, r.Err)
			return nil
		}
	}

	return json.NewEncoder(w).Encode(r)
}
