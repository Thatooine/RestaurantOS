package restaurants

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// RestaurantServiceRESTAdaptor exposes restaurant read operations over a REST API.
type RestaurantServiceRESTAdaptor struct {
	service RestaurantService
}

// NewRestaurantServiceRESTAdaptor returns a new RestaurantServiceRESTAdaptor.
func NewRestaurantServiceRESTAdaptor(service RestaurantService) *RestaurantServiceRESTAdaptor {
	return &RestaurantServiceRESTAdaptor{service: service}
}

// GetMyRestaurant

func (a *RestaurantServiceRESTAdaptor) GetMyRestaurant(w http.ResponseWriter, r *http.Request) {
	claim, ok := authentication.LoginClaimFromContext(r.Context())
	if !ok {
		log.Ctx(r.Context()).Warn().Msg("no login claim in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	resp, err := a.service.GetMyRestaurant(r.Context(), GetMyRestaurantRequest{OwnerID: claim.UserID})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to get my restaurant")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetRestaurantRESTResponse{Restaurant: resp.Restaurant})
}

// GetRestaurant

type GetRestaurantRESTResponse struct {
	Restaurant Restaurant `json:"restaurant"`
}

func (a *RestaurantServiceRESTAdaptor) GetRestaurant(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	resp, err := a.service.GetRestaurant(r.Context(), GetRestaurantRequest{ID: id})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to get restaurant")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetRestaurantRESTResponse{Restaurant: resp.Restaurant})
}

// ListRestaurants

type ListRestaurantsRESTResponse struct {
	Restaurants []Restaurant `json:"restaurants"`
	Total       int64        `json:"total"`
}

func (a *RestaurantServiceRESTAdaptor) ListRestaurants(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.ListRestaurants(r.Context(), ListRestaurantsRequest{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to list restaurants")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListRestaurantsRESTResponse{
		Restaurants: resp.Restaurants,
		Total:       resp.Total,
	})
}

// SearchRestaurants

type SearchRestaurantsRESTResponse struct {
	Restaurants []Restaurant `json:"restaurants"`
	Total       int64        `json:"total"`
}

func (a *RestaurantServiceRESTAdaptor) SearchRestaurants(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	q := query.Get("q")
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.SearchRestaurants(
		r.Context(),
		SearchRestaurantsRequest{
			Query:  q,
			Offset: offset,
			Limit:  limit,
		})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to search restaurants")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SearchRestaurantsRESTResponse{
		Restaurants: resp.Restaurants,
		Total:       resp.Total,
	})
}
