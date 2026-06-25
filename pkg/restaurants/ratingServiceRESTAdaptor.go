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

// RatingServiceRESTAdaptor exposes rating submission and read operations over a REST API.
type RatingServiceRESTAdaptor struct {
	service RatingService
}

// NewRatingServiceRESTAdaptor returns a new RatingServiceRESTAdaptor.
func NewRatingServiceRESTAdaptor(service RatingService) *RatingServiceRESTAdaptor {
	return &RatingServiceRESTAdaptor{service: service}
}

// SubmitRating

type SubmitRatingRESTRequest struct {
	DishID string `json:"dish_id"`
	Score  int    `json:"score"`
	Review string `json:"review"`
}

type SubmitRatingRESTResponse struct {
	Rating Rating `json:"rating"`
}

func (a *RatingServiceRESTAdaptor) SubmitRating(w http.ResponseWriter, r *http.Request) {
	var request SubmitRatingRESTRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to decode submit rating request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	claim, ok := authentication.LoginClaimFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	resp, err := a.service.SubmitRating(r.Context(), SubmitRatingRequest{
		DishID: request.DishID,
		UserID: claim.UserID,
		Score:  request.Score,
		Review: request.Review,
	})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to submit rating")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SubmitRatingRESTResponse{Rating: resp.Rating})
}

// ListRatings

type ListRatingsRESTResponse struct {
	Ratings []Rating `json:"ratings"`
	Total   int64    `json:"total"`
}

func (a *RatingServiceRESTAdaptor) ListRatings(w http.ResponseWriter, r *http.Request) {
	dishID := mux.Vars(r)["id"]
	query := r.URL.Query()
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.ListRatings(r.Context(), ListRatingsRequest{
		DishID: dishID,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to list ratings")
		errs.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListRatingsRESTResponse{
		Ratings: resp.Ratings,
		Total:   resp.Total,
	})
}
