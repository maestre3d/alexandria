package main

import (
	"encoding/json"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/httputil"
	"github.com/gorilla/mux"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"github.com/maestre3d/alexandria/identity-service/internal/infrastructure/dependency"
	"github.com/maestre3d/alexandria/identity-service/internal/interactor"
	"log"
	"net/http"
)

var userUseCase *interactor.UserUseCase
var identityUseCase *interactor.IdentityUseCase

func main() {
	userUC, cleanup, err := dependency.InjectUserUseCase()
	if err != nil {
		panic(err)
	}
	defer cleanup()
	userUseCase = userUC

	identityUC, cleanup, err := dependency.InjectIdentityUseCase()
	if err != nil {
		panic(err)
	}
	defer cleanup()
	identityUseCase = identityUC

	// Root router
	r := mux.NewRouter()

	// Parent routing
	pub := r.PathPrefix(core.PublicAPI).Subrouter()
	// adm := r.PathPrefix(core.AdminAPI).Subrouter()

	// Child routing
	authPublic := pub.PathPrefix("/auth").Subrouter()
	userPublic := pub.PathPrefix("/user").Subrouter()

	//_ = adm.PathPrefix("/auth").Subrouter()
	//_ = adm.PathPrefix("/user").Subrouter()

	authPublic.Path("/signup").Methods(http.MethodPost).HandlerFunc(SignUpHandler)
	authPublic.Path("/signin").Methods(http.MethodPost).HandlerFunc(SignInHandler)
	authPublic.Path("/confirm").Methods(http.MethodPost).HandlerFunc(ConfirmSignUpHandler)

	userPublic.Path("").Methods(http.MethodGet).HandlerFunc(ListUserHandler)
	userPublic.Path("").Methods(http.MethodPost).HandlerFunc(CreateUserHandler)

	log.Print("http server configured")

	panic(http.ListenAndServe(":8080", r))
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	body := domain.UserAggregate{}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		httputil.ResponseErrJSON(w, err)
		return
	}

	err = identityUseCase.SignUp(r.Context(), body)
	if err != nil {
		httputil.ResponseErrJSON(w, err)
		return
	}

	user, err := userUseCase.Create(r.Context(), body)
	if err != nil {
		httputil.ResponseErrJSON(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(user)
}

func ConfirmSignUpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	body := struct {
		Username string `json:"username"`
		Code     string `json:"code"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		httputil.ResponseErrJSON(w, err)
		return
	}

	err = identityUseCase.ConfirmSignUp(r.Context(), body.Username, body.Code)
	if err != nil {
		httputil.ResponseErrJSON(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(&struct {
	}{})
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	body := struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		RefreshToken string `json:"refresh_token"`
		DeviceKey    string `json:"device_key"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		httputil.ResponseErrJSON(w, err)
		return
	}

	token, err := identityUseCase.SignIn(r.Context(), body.Username, body.Password, body.RefreshToken, body.DeviceKey)
	if err != nil {
		httputil.ResponseErrJSON(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(token)
}

func ListUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	users, nt, err := userUseCase.List(r.Context(), core.NewPaginationParams("", "2"), nil)
	if err != nil {
		httputil.ResponseErrJSON(w, err)
		return
	}

	res := struct {
		Users     []*domain.User `json:"users"`
		NextToken string         `json:"next_page_token"`
	}{users, nt}

	_ = json.NewEncoder(w).Encode(&res)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	body := domain.UserAggregate{}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		httputil.ResponseErrJSON(w, err)
		return
	}

	user, err := userUseCase.Create(r.Context(), body)
	if err != nil {
		httputil.ResponseErrJSON(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(user)
}
