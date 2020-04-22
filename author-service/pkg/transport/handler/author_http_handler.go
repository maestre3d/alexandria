package handler

import (
	"context"
	"encoding/json"
	"errors"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	kitoc "github.com/go-kit/kit/tracing/opencensus"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/exception"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/helper"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strings"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/author-service/pkg/author/action"
)

type AuthorHandler struct {
	service      service.IAuthorService
	logger       log.Logger
	duration     *kitprometheus.Summary
	tracer       stdopentracing.Tracer
	zipkinTracer *stdzipkin.Tracer
	options      []httptransport.ServerOption
}

func NewAuthorHandler(svc service.IAuthorService, logger log.Logger, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) *AuthorHandler {
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
		httptransport.ServerErrorEncoder(helper.ErrorEncoder),
		kitoc.HTTPServerTrace(),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	if zipkinTracer != nil {
		// Zipkin HTTP Server Trace can either be instantiated per endpoint with a
		// provided operation name or a global tracing service can be instantiated
		// without an operation name and fed to each Go kit endpoint as ServerOption.
		// In the latter case, the operation name will be the endpoint's http method.
		// We demonstrate a global tracing service here.
		options = append(options, zipkin.HTTPServerTrace(zipkinTracer))
	}

	return &AuthorHandler{svc, logger, duration, tracer, zipkinTracer, options}
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

/* Decoders/Encoders */

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.CreateRequest{
		FirstName:   r.PostFormValue("first_name"),
		LastName:    r.PostFormValue("last_name"),
		DisplayName: r.PostFormValue("display_name"),
		BirthDate:   r.PostFormValue("birth_date"),
	}, nil
}

func decodeListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.ListRequest{
		PageToken: r.URL.Query().Get("page_token"),
		PageSize:  r.URL.Query().Get("page_size"),
		FilterParams: util.FilterParams{
			"query":     r.URL.Query().Get("query"),
			"timestamp": r.URL.Query().Get("timestamp"),
		},
	}, nil
}

func decodeGetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	return action.GetRequest{params["id"]}, nil
}

func decodeUpdateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.UpdateRequest{
		ID:          mux.Vars(r)["id"],
		FirstName:   r.PostFormValue("first_name"),
		LastName:    r.PostFormValue("last_name"),
		DisplayName: r.PostFormValue("display_name"),
		BirthDate:   r.PostFormValue("birth_date"),
	}, nil
}

func decodeDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	params := mux.Vars(r)
	return action.DeleteRequest{params["id"]}, nil
}

func encodeCreateResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.CreateResponse)
	if ok {
		if r.Err != nil {
			if errors.Is(r.Err, exception.InvalidFieldFormat) || errors.Is(r.Err, exception.InvalidFieldRange) || errors.Is(r.Err, exception.RequiredField) {
				errDesc := strings.Split(r.Err.Error(), ":")
				w.WriteHeader(http.StatusBadRequest)
				return json.NewEncoder(w).Encode(shared.Error{errDesc[len(errDesc)-1]})
			} else if errors.Is(r.Err, exception.EntityExists) {
				w.WriteHeader(http.StatusConflict)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			return json.NewEncoder(w).Encode(shared.Error{r.Err.Error()})
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeListResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.ListResponse)
	if ok {
		if r.Err != nil {
			if errors.Is(r.Err, exception.InvalidID) {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			return json.NewEncoder(w).Encode(shared.Error{r.Err.Error()})
		} else if r.Err == nil && len(r.Authors) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return json.NewEncoder(w).Encode(shared.Error{exception.EntitiesNotFound.Error()})
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeGetResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.GetResponse)
	if ok {
		if r.Err != nil {
			if errors.Is(r.Err, exception.InvalidID) {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			return json.NewEncoder(w).Encode(shared.Error{r.Err.Error()})
		} else if r.Err == nil && r.Author == nil {
			w.WriteHeader(http.StatusNotFound)
			return json.NewEncoder(w).Encode(shared.Error{exception.EntityNotFound.Error()})
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeUpdateResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.UpdateResponse)
	if ok {
		if r.Err != nil {
			if errors.Is(r.Err, exception.InvalidFieldFormat) || errors.Is(r.Err, exception.InvalidFieldRange) {
				errDesc := strings.Split(r.Err.Error(), ":")
				w.WriteHeader(http.StatusBadRequest)
				return json.NewEncoder(w).Encode(shared.Error{errDesc[len(errDesc)-1]})
			} else if errors.Is(r.Err, exception.InvalidID) || errors.Is(r.Err, exception.EmptyBody) {
				w.WriteHeader(http.StatusBadRequest)
			} else if errors.Is(r.Err, exception.EntityExists) {
				w.WriteHeader(http.StatusConflict)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			return json.NewEncoder(w).Encode(shared.Error{r.Err.Error()})
		}
	}

	return json.NewEncoder(w).Encode(r)
}

func encodeDeleteResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.DeleteResponse)
	if ok {
		if r.Err != nil {
			if errors.Is(r.Err, exception.InvalidID) {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			return json.NewEncoder(w).Encode(shared.Error{r.Err.Error()})
		}
	}

	return json.NewEncoder(w).Encode(r)
}
