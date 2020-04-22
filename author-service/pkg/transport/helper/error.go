package helper

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/exception"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"net/http"
)

func ErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(ErrToCode(err))
	json.NewEncoder(w).Encode(shared.Error{Message: err.Error()})
}

func ErrorDecoder(r *http.Response) error {
	var w shared.Error
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Message)
}

func ErrToCode(err error) int {
	switch err {
	case exception.InvalidFieldRange, exception.EmptyBody, exception.InvalidFieldFormat, exception.InvalidID, exception.RequiredField:
		return http.StatusBadRequest
	case exception.EntityExists:
		return http.StatusConflict
	case exception.EntitiesNotFound, exception.EntityNotFound:
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

