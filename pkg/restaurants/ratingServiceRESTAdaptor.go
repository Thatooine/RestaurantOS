package restaurants

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	"github.com/gin-gonic/gin"
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

func (a *RatingServiceRESTAdaptor) SubmitRating(c *gin.Context) {
	ctx := c.Request.Context()
	var request SubmitRatingRESTRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to decode submit rating request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claim, ok := authentication.LoginClaimFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Preserve the existing API contract: dish_id in the request body is the
	// effective ID even though the route also contains an :id parameter.
	resp, err := a.service.SubmitRating(ctx, SubmitRatingRequest{
		DishID: request.DishID,
		UserID: claim.UserID,
		Score:  request.Score,
		Review: request.Review,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to submit rating")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusCreated, SubmitRatingRESTResponse{Rating: resp.Rating})
}

// ListRatings

type ListRatingsRESTResponse struct {
	Ratings []Rating `json:"ratings"`
	Total   int64    `json:"total"`
}

func (a *RatingServiceRESTAdaptor) ListRatings(c *gin.Context) {
	ctx := c.Request.Context()
	dishID := c.Param("id")
	query := c.Request.URL.Query()
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.ListRatings(ctx, ListRatingsRequest{
		DishID: dishID,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to list ratings")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusOK, ListRatingsRESTResponse{
		Ratings: resp.Ratings,
		Total:   resp.Total,
	})
}
