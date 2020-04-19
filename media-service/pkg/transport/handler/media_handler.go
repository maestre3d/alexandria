package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/exception"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/media-service/pkg/media/service"
	"github.com/maestre3d/alexandria/media-service/pkg/shared"
	"net/http"
	"strings"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/media-service/pkg/media/action"
)

type MediaHandler struct {
	service service.IMediaService
	logger  log.Logger
}

func NewMediaHandler(svc service.IMediaService, logger log.Logger) *MediaHandler {
	return &MediaHandler{svc, logger}
}

func (h *MediaHandler) Create() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeCreateMediaEndpoint(h.service, h.logger),
		decodeCreateRequest,
		encodeCreateRequest,
	)
}

func (h *MediaHandler) List() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeListMediaEndpoint(h.service, h.logger),
		decodeListRequest,
		encodeListResponse,
	)
}

func (h *MediaHandler) Get() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeGetMediaEndpoint(h.service, h.logger),
		decodeGetRequest,
		encodeGetResponse,
	)
}

func (h *MediaHandler) Update() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeUpdateMediaEndpoint(h.service, h.logger),
		decodeUpdateRequest,
		encodeUpdateResponse,
	)
}

func (h *MediaHandler) Delete() *httptransport.Server {
	return httptransport.NewServer(
		action.MakeDeleteMediaEndpoint(h.service, h.logger),
		decodeDeleteRequest,
		encodeDeleteResponse,
	)
}

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.CreateRequest{
		Title:       r.PostFormValue("title"),
		DisplayName: r.PostFormValue("display_name"),
		Description: r.PostFormValue("description"),
		UserID:      r.PostFormValue("user_id"),
		AuthorID:    r.PostFormValue("author_id"),
		PublishDate: r.PostFormValue("publish_date"),
		MediaType:   r.PostFormValue("media_type"),
	}, nil
}

func decodeListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return action.ListRequest{
		PageToken: r.URL.Query().Get("page_token"),
		PageSize:  r.URL.Query().Get("page_size"),
		FilterParams: util.FilterParams{
			"query":     r.URL.Query().Get("search_query"),
			"timestamp": r.URL.Query().Get("timestamp"),
			"author":    r.URL.Query().Get("author"),
			"user":      r.URL.Query().Get("user"),
			"media":     r.URL.Query().Get("media"),
			"title":     r.URL.Query().Get("title"),
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
		Title:       r.PostFormValue("title"),
		DisplayName: r.PostFormValue("display_name"),
		Description: r.PostFormValue("description"),
		UserID:      r.PostFormValue("user_id"),
		AuthorID:    r.PostFormValue("author_id"),
		PublishDate: r.PostFormValue("publish_date"),
		MediaType:   r.PostFormValue("media_type"),
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
		} else if r.Err == nil && len(r.Media) == 0 {
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
		} else if r.Err == nil && r.Media == nil {
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