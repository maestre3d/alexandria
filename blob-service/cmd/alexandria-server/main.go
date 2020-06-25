package main

import (
	"context"
	"encoding/json"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/exception"
	"github.com/alexandria-oss/core/httputil"
	"github.com/alexandria-oss/core/logger"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	"github.com/maestre3d/alexandria/blob-service/internal/infrastructure"
	"github.com/maestre3d/alexandria/blob-service/internal/interactor"
	"math"
	"net/http"
)

func main() {
	ctx := context.Background()
	zapLog := logger.NewZapLogger()
	cfg, err := config.NewKernel(ctx)
	if err != nil {
		panic(err)
	}
	repo := infrastructure.NewBlobDynamoRepository(zapLog, cfg)
	storage := infrastructure.NewBlobS3Storage(zapLog)
	blobInter := interactor.NewBlob(zapLog, repo, storage)

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

		b, err := blobInter.Store(r.Context(), mux.Vars(r)["id"], domain.Media, h)
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
	r.PathPrefix(core.PrivateAPI + "/blob/{id}").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		b, err := blobInter.Get(r.Context(), id)
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
	r.PathPrefix(core.PrivateAPI + "/blob/{id}").Methods(http.MethodDelete).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json; utf-8")
		id := mux.Vars(r)["id"]
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(&httputil.GenericResponse{
				Message: "missing required request field id",
				Code:    http.StatusBadRequest,
			})
		}

		err := blobInter.Delete(r.Context(), id)
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
