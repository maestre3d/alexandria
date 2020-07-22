package transport

import (
	"encoding/json"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/httputil"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"github.com/maestre3d/alexandria/category-service/pkg/service"
	"github.com/maestre3d/alexandria/category-service/pkg/transport/observability"
	"net/http"
)

type CategoryHTTP struct {
	svc service.Category
}

func NewCategoryHTTP(svc service.Category) *CategoryHTTP {
	return &CategoryHTTP{
		svc: svc,
	}
}

func (t *CategoryHTTP) SetRoutes(public, private, admin *mux.Router) {
	// Using OpenCensus middleware for distributed tracing
	public.Path("/category").Methods(http.MethodGet).Handler(observability.Trace(t.list, true))
	public.Path("/category/{id}").Methods(http.MethodGet).Handler(observability.Trace(t.get, true))

	private.Path("/category").Methods(http.MethodPost).Handler(observability.Trace(t.create, false))
	private.Path("/category/{id}").Methods(http.MethodPatch, http.MethodPut).Handler(observability.Trace(t.update, false))
	private.Path("/category/{id}").Methods(http.MethodDelete).Handler(observability.Trace(t.delete, false))

	admin.Path("/category/{id}/restore").Methods(http.MethodGet).Handler(observability.Trace(t.restore, false))
	admin.Path("/category/{id}").Methods(http.MethodDelete).Handler(observability.Trace(t.hardDelete, false))
}

func (t *CategoryHTTP) create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	category, err := t.svc.Create(r.Context(), r.PostForm.Get("name"))
	if err != nil {
		httputil.ResponseErrJSON(r.Context(), err, w)
		return
	}

	_ = json.NewEncoder(w).Encode(category)
}

func (t *CategoryHTTP) get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	category, err := t.svc.Get(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		httputil.ResponseErrJSON(r.Context(), err, w)
		return
	}

	_ = json.NewEncoder(w).Encode(category)
}

func (t *CategoryHTTP) list(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	filter := core.FilterParams{
		"query": r.URL.Query().Get("query"),
	}

	categories, nextToken, err := t.svc.List(r.Context(), r.URL.Query().Get("next_token"),
		r.URL.Query().Get("limit"), filter)
	if err != nil {
		httputil.ResponseErrJSON(r.Context(), err, w)
		return
	}

	_ = json.NewEncoder(w).Encode(&struct {
		Categories []*domain.Category `json:"categories"`
		NextToken  string             `json:"next_token"`
	}{
		Categories: categories,
		NextToken:  nextToken,
	})
}

func (t *CategoryHTTP) update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	category, err := t.svc.Update(r.Context(), mux.Vars(r)["id"], r.PostForm.Get("name"))
	if err != nil {
		httputil.ResponseErrJSON(r.Context(), err, w)
		return
	}

	_ = json.NewEncoder(w).Encode(category)
}

func (t *CategoryHTTP) delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	err := t.svc.Delete(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		httputil.ResponseErrJSON(r.Context(), err, w)
		return
	}

	_ = json.NewEncoder(w).Encode(&struct{}{})
}

func (t *CategoryHTTP) restore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	err := t.svc.Restore(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		httputil.ResponseErrJSON(r.Context(), err, w)
		return
	}

	_ = json.NewEncoder(w).Encode(&struct{}{})
}

func (t *CategoryHTTP) hardDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	err := t.svc.HardDelete(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		httputil.ResponseErrJSON(r.Context(), err, w)
		return
	}

	_ = json.NewEncoder(w).Encode(&struct{}{})
}
