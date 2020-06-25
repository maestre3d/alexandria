package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"github.com/alexandria-oss/core/httputil"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/blob-service/internal/dependency"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	"math"
	"net/http"
	"strings"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dependency.Ctx = ctx
	blobInter, cleanup, err := dependency.InjectBlobUseCase()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	r := mux.NewRouter()
	// Store
	r.PathPrefix(core.PrivateAPI + "/blob/media/{id}").Methods(http.MethodPost).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json; utf-8")
		f, h, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(&httputil.GenericResponse{
				Message: err.Error(),
				Code:    http.StatusInternalServerError,
			})
			return
		}
		defer f.Close()

		contentSlice := strings.Split(h.Header.Get("Content-Type"), "/")
		if len(contentSlice) <= 1 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(&httputil.GenericResponse{
				Message: fmt.Sprintf(exception.InvalidFieldFormatString,
					"file", "invalid file extension"),
				Code: http.StatusBadRequest,
			})
			return
		}

		// go's file.size is in bytes
		var maxSize int64
		maxSize = 8192 * (int64(math.Pow(1024, 2)))

		if h.Size > maxSize {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(&httputil.GenericResponse{
				Message: "file size off limits",
				Code:    http.StatusBadRequest,
			})
			return
		}

		b, err := blobInter.Store(r.Context(), domain.BlobAggregate{
			RootID:    mux.Vars(r)["id"],
			Service:   domain.Media,
			BlobType:  contentSlice[0],
			Extension: contentSlice[1],
			Size:      h.Size,
			Content:   f,
		})
		if err != nil {
			code := httputil.ErrorToCode(err)
			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(&httputil.GenericResponse{
				Message: exception.GetErrorDescription(err),
				Code:    code,
			})
			return
		}

		_ = json.NewEncoder(w).Encode(b)
	})

	// Get
	r.PathPrefix(core.PrivateAPI + "/blob/media/{id}").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json; utf-8")
		id := mux.Vars(r)["id"]
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(&httputil.GenericResponse{
				Message: "missing required request field id",
				Code:    http.StatusBadRequest,
			})
		}

		// Add sort key for NoSQL databases
		/*
			if sortKey := r.URL.Query().Get("sort_key"); sortKey != "" {
				id += ": " + sortKey
			}*/

		b, err := blobInter.Get(r.Context(), id, domain.Media)
		if err != nil {
			code := httputil.ErrorToCode(err)
			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(&httputil.GenericResponse{
				Message: exception.GetErrorDescription(err),
				Code:    code,
			})
			return
		}

		_ = json.NewEncoder(w).Encode(b)
	})

	// Delete
	r.PathPrefix(core.PrivateAPI + "/blob/media/{id}").Methods(http.MethodDelete).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json; utf-8")
		id := mux.Vars(r)["id"]
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(&httputil.GenericResponse{
				Message: "missing required request field id",
				Code:    http.StatusBadRequest,
			})
		}

		err := blobInter.Delete(r.Context(), id, domain.Media)
		if err != nil {
			code := httputil.ErrorToCode(err)
			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(&httputil.GenericResponse{
				Message: exception.GetErrorDescription(err),
				Code:    code,
			})
			return
		}

		_ = json.NewEncoder(w).Encode(struct{}{})
	})
	panic(http.ListenAndServe(":8080", r))
}
