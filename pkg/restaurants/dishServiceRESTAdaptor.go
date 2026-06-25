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

// DishServiceRESTAdaptor exposes dish read and write operations over a REST API.
type DishServiceRESTAdaptor struct {
	service DishService
}

// NewDishServiceRESTAdaptor returns a new DishServiceRESTAdaptor.
func NewDishServiceRESTAdaptor(service DishService) *DishServiceRESTAdaptor {
	return &DishServiceRESTAdaptor{service: service}
}

// GetDish

type GetDishRESTResponse struct {
	Dish Dish `json:"dish"`
}

func (a *DishServiceRESTAdaptor) GetDish(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	resp, err := a.service.GetDish(r.Context(), GetDishRequest{ID: id})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to get dish")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetDishRESTResponse{Dish: resp.Dish})
}

// ListDishes

type ListDishesRESTResponse struct {
	Dishes []Dish `json:"dishes"`
	Total  int64  `json:"total"`
}

func (a *DishServiceRESTAdaptor) ListDishes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	restaurantID := query.Get("restaurant_id")
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.ListDishes(r.Context(), ListDishesRequest{
		RestaurantID: restaurantID,
		Offset:       offset,
		Limit:        limit,
	})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to list dishes")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListDishesRESTResponse{
		Dishes: resp.Dishes,
		Total:  resp.Total,
	})
}

// SearchDishes

type SearchDishesRESTResponse struct {
	Dishes []Dish `json:"dishes"`
	Total  int64  `json:"total"`
}

func (a *DishServiceRESTAdaptor) SearchDishes(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	q := query.Get("q")
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.SearchDishes(
		r.Context(),
		SearchDishesRequest{
			Query:  q,
			Offset: offset,
			Limit:  limit,
		})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to search dishes")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SearchDishesRESTResponse{
		Dishes: resp.Dishes,
		Total:  resp.Total,
	})
}

// CreateDish

type CreateDishRESTRequest struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	RestaurantID string  `json:"restaurant_id"`
	Image        string  `json:"image"`
}

type CreateDishRESTResponse struct {
	Dish Dish `json:"dish"`
}

func (a *DishServiceRESTAdaptor) CreateDish(w http.ResponseWriter, r *http.Request) {
	claim, ok := authentication.LoginClaimFromContext(r.Context())
	if !ok {
		log.Ctx(r.Context()).Warn().Msg("no login claim in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	var request CreateDishRESTRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to decode create dish request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp, err := a.service.CreateDish(r.Context(), CreateDishRequest{
		UserID:       claim.UserID,
		Name:         request.Name,
		Description:  request.Description,
		Price:        request.Price,
		RestaurantID: request.RestaurantID,
		Image:        request.Image,
	})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to create dish")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateDishRESTResponse{Dish: resp.Dish})
}

// UpdateDish

type UpdateDishRESTRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
}

type UpdateDishRESTResponse struct {
	Dish Dish `json:"dish"`
}

func (a *DishServiceRESTAdaptor) UpdateDish(w http.ResponseWriter, r *http.Request) {
	claim, ok := authentication.LoginClaimFromContext(r.Context())
	if !ok {
		log.Ctx(r.Context()).Warn().Msg("no login claim in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	id := mux.Vars(r)["id"]

	var request UpdateDishRESTRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to decode update dish request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp, err := a.service.UpdateDish(r.Context(), UpdateDishRequest{
		UserID:      claim.UserID,
		ID:          id,
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
		Image:       request.Image,
	})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to update dish")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UpdateDishRESTResponse{Dish: resp.Dish})
}

// DeleteDish

func (a *DishServiceRESTAdaptor) DeleteDish(w http.ResponseWriter, r *http.Request) {
	claim, ok := authentication.LoginClaimFromContext(r.Context())
	if !ok {
		log.Ctx(r.Context()).Warn().Msg("no login claim in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	id := mux.Vars(r)["id"]

	if err := a.service.DeleteDish(r.Context(), DeleteDishRequest{UserID: claim.UserID, ID: id}); err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to delete dish")
		errs.WriteHTTPError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
