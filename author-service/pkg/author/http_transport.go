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
	"io"
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
		encodeGetResponse,
	)

	updateHandler := httptransport.NewServer(
		action.MakeUpdateAuthorEndpoint(svc, logger),
		decodeUpdateRequest,
		encodeUpdateResponse,
	)

	deleteHandler := httptransport.NewServer(
		action.MakeDeleteAuthorEndpoint(svc, logger),
		decodeDeleteRequest,
		encodeDeleteResponse,
	)

	r := mux.NewRouter()
	r.Methods(http.MethodOptions)

	apiRouter := r.PathPrefix("/v1").Subrouter()
	apiRouter.PathPrefix("/health").Methods(http.MethodGet).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		io.WriteString(writer, `{"alive":true}`)
	})
	apiRouter.PathPrefix("/metrics").Methods(http.MethodGet).Handler(promhttp.Handler())

	authorRouter := apiRouter.PathPrefix("/author").Subrouter()
	authorRouter.Path("").Methods(http.MethodPost).Handler(createHandler)
	authorRouter.Path("").Methods(http.MethodGet).Handler(listHandler)
	authorRouter.Path("/").Methods(http.MethodPost).Handler(createHandler)
	authorRouter.Path("/").Methods(http.MethodGet).Handler(listHandler)

	authorRouter.Path("/{id}").Methods(http.MethodGet).Handler(getHandler)
	authorRouter.Path("/{id}").Methods(http.MethodPatch, http.MethodPut).Handler(updateHandler)
	authorRouter.Path("/{id}").Methods(http.MethodDelete).Handler(deleteHandler)

	r.Use(mux.CORSMethodMiddleware(r))
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
				return json.NewEncoder(w).Encode(shared.Error{errDesc[len(errDesc) - 1]})
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