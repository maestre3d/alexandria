package transport

import (
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/event-telemetry-service/internal/telemetry/interactor"
	"github.com/maestre3d/alexandria/event-telemetry-service/pkg/shared"
	"net/http"
	"strings"
	"time"
)

var usecase *interactor.EventUseCase

func NewHTTPServer(eventUseCase *interactor.EventUseCase, cfg *config.KernelConfiguration) *http.Server {
	usecase = eventUseCase

	router := mux.NewRouter().PathPrefix("/v1/event").Subrouter()

	router.Path("").Methods(http.MethodGet).HandlerFunc(listEventHandler)
	router.Path("/").Methods(http.MethodGet).HandlerFunc(listEventHandler)
	router.Path("/{id}").Methods(http.MethodGet).HandlerFunc(getEventHandler)

	return &http.Server{
		Addr:              cfg.TransportConfig.HTTPHost + fmt.Sprintf(":%d", cfg.TransportConfig.HTTPPort),
		Handler:           router,
		TLSConfig:         nil,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       15 * time.Second,
		MaxHeaderBytes:    4096,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
		BaseContext:       nil,
		ConnContext:       nil,
	}
}

func listEventHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	filter := core.FilterParams{
		"query": r.URL.Query().Get("query"),
	}

	events, nextToken, err := usecase.List(r.URL.Query().Get("page_token"), r.URL.Query().Get("page_size"), filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(w).Encode(&shared.Error{err.Error()})
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		return
	} else if len(events) == 0 {
		w.WriteHeader(http.StatusNotFound)
		err = json.NewEncoder(w).Encode(&shared.Error{exception.EntityNotFound.Error()})
		if err != nil {
			w.Write([]byte(exception.EntityNotFound.Error()))
		}
		return
	}

	payload := struct {
		Events    []*eventbus.Event `json:"events"`
		NextToken string            `json:"next_page_token"`
	}{
		events,
		nextToken,
	}

	err = json.NewEncoder(w).Encode(payload)
	if err != nil {
		w.Write(nil)
	}
}

func getEventHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.Split(r.URL.Path, "/")

	event, err := usecase.Get(path[len(path)-1])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(w).Encode(&shared.Error{err.Error()})
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		return
	} else if event == nil {
		w.WriteHeader(http.StatusNotFound)
		err = json.NewEncoder(w).Encode(&shared.Error{exception.EntityNotFound.Error()})
		if err != nil {
			w.Write([]byte(exception.EntityNotFound.Error()))
		}
		return
	}

	payload := struct {
		Event *eventbus.Event `json:"event"`
	}{event}

	err = json.NewEncoder(w).Encode(payload)
	if err != nil {
		w.Write(nil)
	}
}
