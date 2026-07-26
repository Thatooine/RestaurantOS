package restaurants

import (
	"net/http"
	"strconv"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	"github.com/gin-gonic/gin"
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

func (a *RestaurantServiceRESTAdaptor) GetMyRestaurant(c *gin.Context) {
	ctx := c.Request.Context()
	claim, ok := authentication.LoginClaimFromGinContext(c)
	if !ok {
		log.Ctx(ctx).Warn().Msg("no login claim in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := a.service.GetMyRestaurant(ctx, GetMyRestaurantRequest{OwnerID: claim.UserID})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get my restaurant")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusOK, GetRestaurantRESTResponse{Restaurant: resp.Restaurant})
}

// GetRestaurant

type GetRestaurantRESTResponse struct {
	Restaurant Restaurant `json:"restaurant"`
}

func (a *RestaurantServiceRESTAdaptor) GetRestaurant(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	resp, err := a.service.GetRestaurant(ctx, GetRestaurantRequest{ID: id})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get restaurant")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusOK, GetRestaurantRESTResponse{Restaurant: resp.Restaurant})
}

// ListRestaurants

type ListRestaurantsRESTResponse struct {
	Restaurants []Restaurant `json:"restaurants"`
	Total       int64        `json:"total"`
}

func (a *RestaurantServiceRESTAdaptor) ListRestaurants(c *gin.Context) {
	ctx := c.Request.Context()
	query := c.Request.URL.Query()
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.ListRestaurants(ctx, ListRestaurantsRequest{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to list restaurants")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusOK, ListRestaurantsRESTResponse{
		Restaurants: resp.Restaurants,
		Total:       resp.Total,
	})
}

// SearchRestaurants

type SearchRestaurantsRESTResponse struct {
	Restaurants []Restaurant `json:"restaurants"`
	Total       int64        `json:"total"`
}

func (a *RestaurantServiceRESTAdaptor) SearchRestaurants(c *gin.Context) {
	ctx := c.Request.Context()
	query := c.Request.URL.Query()
	q := query.Get("q")
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.SearchRestaurants(
		ctx,
		SearchRestaurantsRequest{
			Query:  q,
			Offset: offset,
			Limit:  limit,
		})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to search restaurants")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusOK, SearchRestaurantsRESTResponse{
		Restaurants: resp.Restaurants,
		Total:       resp.Total,
	})
}
