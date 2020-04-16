package author

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/exception"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strings"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/author-service/pkg/author/action"
)

func NewTransportHTTP(svc service.IAuthorService, logger log.Logger) *mux.Router {
	// TODO: Add metrics with OpenCensus and Prometheus/Zipkin
	createHandler := httptransport.NewServer(
		action.MakeCreateAuthorEndpoint(svc, logger),
		decodeCreateRequest,
		encodeCreateRequest,
	)

	listHandler := httptransport.NewServer(
		action.MakeListAuthorEndpoint(svc, logger),
		decodeListRequest,
		encodeListResponse,
	)

	getHandler := httptransport.NewServer(
		action.MakeGetAuthorEndpoint(svc, logger),
		decodeGetRequest,
		httptransport.EncodeJSONResponse,
	)

	updateHandler := httptransport.NewServer(
		action.MakeUpdateAuthorEndpoint(svc, logger),
		decodeUpdateRequest,
		httptransport.EncodeJSONResponse,
	)

	deleteHandler := httptransport.NewServer(
		action.MakeDeleteAuthorEndpoint(svc, logger),
		decodeDeleteRequest,
		httptransport.EncodeJSONResponse,
	)

	r := mux.NewRouter()
	apiRouter := r.PathPrefix("/v1").Subrouter()
	apiRouter.PathPrefix("/metrics").Methods("GET").Handler(promhttp.Handler())

	authorRouter := apiRouter.PathPrefix("/author").Subrouter()
	authorRouter.Methods("POST").Handler(createHandler)
	authorRouter.Methods("GET").Handler(listHandler)

	authorDetailR := authorRouter.PathPrefix("/{author}").Subrouter()
	authorDetailR.Methods("GET").Handler(getHandler)
	authorDetailR.Methods("PATCH").Handler(updateHandler)
	authorDetailR.Methods("PUT").Handler(updateHandler)
	authorDetailR.Methods("DELETE").Handler(deleteHandler)

	return r
}

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
		PageToken:    r.URL.Query().Get("page_token"),
		PageSize:     r.URL.Query().Get("page_size"),
		FilterParams: util.FilterParams{
			"query":r.URL.Query().Get("search_query"),
			"timestamp":r.URL.Query().Get("timestamp"),
		},
	}, nil
}

func encodeCreateRequest(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r, ok := response.(action.CreateResponse)
	if ok {
		if r.Err != nil {
			if errors.Is(r.Err, exception.InvalidFieldFormat) || errors.Is(r.Err, exception.InvalidFieldRange) || errors.Is(r.Err, exception.RequiredField) {
				errDesc := strings.Split(r.Err.Error(), ":")
				w.WriteHeader(http.StatusBadRequest)
				return json.NewEncoder(w).Encode(shared.Error{errDesc[len(errDesc) - 1]})
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

func encodeListResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
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

func decodeGetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request action.GetRequest
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		return nil, err
	}

	return request, nil
}

func decodeUpdateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request action.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		return nil, err
	}

	return request, nil
}

func decodeDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request action.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		return nil, err
	}

	return request, nil
}