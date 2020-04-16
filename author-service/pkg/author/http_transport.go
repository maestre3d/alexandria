package author

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/exception"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"net/http"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/author-service/pkg/author/action"
)

func NewTransportHTTP(svc service.IAuthorService, logger log.Logger) *mux.Router {
	// logger := log.NewLogfmtLogger(os.Stderr)

	// TODO: Add metrics with OpenCensus and Prometheus/Zipkin
	createHandler := httptransport.NewServer(
		action.MakeCreateAuthorEndpoint(svc, logger),
		decodeCreateRequest,
		httptransport.EncodeJSONResponse,
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
	// r.PathPrefix("/metrics").Methods("GET").Handler(promhttp.Handler())
	r = r.PathPrefix("/author").Subrouter()
	r.Methods("POST").Handler(createHandler)
	r.Methods("GET").Handler(listHandler)
	r.PathPrefix("/{author-id}").Methods("GET").Handler(getHandler)
	r.PathPrefix("/{author-id}").Methods("PATCH").Handler(updateHandler)
	r.PathPrefix("/{author-id}").Methods("PUT").Handler(updateHandler)
	r.PathPrefix("/{author-id}").Methods("DELETE").Handler(deleteHandler)

	return r
}

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request action.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}

	return request, nil
}

func decodeListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	fmt.Println("QUERY:", r.URL.Query())
	var request action.ListRequest

	request.PageToken = r.URL.Query().Get("page_token")
	request.PageSize = r.URL.Query().Get("page_size")
	request.FilterParams = util.FilterParams{
		"query":r.URL.Query().Get("search_query"),
		"timestamp":r.URL.Query().Get("timestamp"),
	}

	return request, nil
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