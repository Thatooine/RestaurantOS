package users

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// UserServiceRESTAdaptor exposes user read operations over a REST API.
type UserServiceRESTAdaptor struct {
	service UserService
}

// NewUserServiceRESTAdaptor returns a new UserServiceRESTAdaptor.
func NewUserServiceRESTAdaptor(service UserService) *UserServiceRESTAdaptor {
	return &UserServiceRESTAdaptor{service: service}
}

// GetUser

type GetUserRESTResponse struct {
	User User `json:"user"`
}

func (a *UserServiceRESTAdaptor) GetUser(w http.ResponseWriter, r *http.Request) {
	email := mux.Vars(r)["email"]

	resp, err := a.service.GetUser(r.Context(), GetUserRequest{Email: email})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to get user")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetUserRESTResponse{User: resp.User})
}

// ListUsers

type ListUsersRESTResponse struct {
	Users []User `json:"users"`
	Total int64  `json:"total"`
}

func (a *UserServiceRESTAdaptor) ListUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.ListUsers(r.Context(), ListUsersRequest{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to list users")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListUsersRESTResponse{
		Users: resp.Users,
		Total: resp.Total,
	})
}

// SearchUsers

type SearchUsersRESTResponse struct {
	Users []User `json:"users"`
	Total int64  `json:"total"`
}

func (a *UserServiceRESTAdaptor) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	q := query.Get("q")
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.SearchUsers(
		r.Context(),
		SearchUsersRequest{
			Query:  q,
			Offset: offset,
			Limit:  limit,
		})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to search users")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SearchUsersRESTResponse{
		Users: resp.Users,
		Total: resp.Total,
	})
}
