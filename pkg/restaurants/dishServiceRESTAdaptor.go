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

func (a *DishServiceRESTAdaptor) GetDish(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	resp, err := a.service.GetDish(ctx, GetDishRequest{ID: id})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get dish")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusOK, GetDishRESTResponse{Dish: resp.Dish})
}

// ListDishes

type ListDishesRESTResponse struct {
	Dishes []Dish `json:"dishes"`
	Total  int64  `json:"total"`
}

func (a *DishServiceRESTAdaptor) ListDishes(c *gin.Context) {
	ctx := c.Request.Context()
	query := c.Request.URL.Query()
	restaurantID := query.Get("restaurant_id")
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.ListDishes(ctx, ListDishesRequest{
		RestaurantID: restaurantID,
		Offset:       offset,
		Limit:        limit,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to list dishes")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusOK, ListDishesRESTResponse{
		Dishes: resp.Dishes,
		Total:  resp.Total,
	})
}

// SearchDishes

type SearchDishesRESTResponse struct {
	Dishes []Dish `json:"dishes"`
	Total  int64  `json:"total"`
}

func (a *DishServiceRESTAdaptor) SearchDishes(c *gin.Context) {
	ctx := c.Request.Context()
	query := c.Request.URL.Query()
	q := query.Get("q")
	offset, _ := strconv.Atoi(query.Get("offset"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	if limit == 0 {
		limit = 20
	}

	resp, err := a.service.SearchDishes(
		ctx,
		SearchDishesRequest{
			Query:  q,
			Offset: offset,
			Limit:  limit,
		})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to search dishes")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusOK, SearchDishesRESTResponse{
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

func (a *DishServiceRESTAdaptor) CreateDish(c *gin.Context) {
	ctx := c.Request.Context()
	claim, ok := authentication.LoginClaimFromGinContext(c)
	if !ok {
		log.Ctx(ctx).Warn().Msg("no login claim in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var request CreateDishRESTRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to decode create dish request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := a.service.CreateDish(ctx, CreateDishRequest{
		UserID:       claim.UserID,
		Name:         request.Name,
		Description:  request.Description,
		Price:        request.Price,
		RestaurantID: request.RestaurantID,
		Image:        request.Image,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to create dish")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusCreated, CreateDishRESTResponse{Dish: resp.Dish})
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

func (a *DishServiceRESTAdaptor) UpdateDish(c *gin.Context) {
	ctx := c.Request.Context()
	claim, ok := authentication.LoginClaimFromGinContext(c)
	if !ok {
		log.Ctx(ctx).Warn().Msg("no login claim in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")

	var request UpdateDishRESTRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to decode update dish request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := a.service.UpdateDish(ctx, UpdateDishRequest{
		UserID:      claim.UserID,
		ID:          id,
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
		Image:       request.Image,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to update dish")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusOK, UpdateDishRESTResponse{Dish: resp.Dish})
}

// DeleteDish

func (a *DishServiceRESTAdaptor) DeleteDish(c *gin.Context) {
	ctx := c.Request.Context()
	claim, ok := authentication.LoginClaimFromGinContext(c)
	if !ok {
		log.Ctx(ctx).Warn().Msg("no login claim in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")

	if err := a.service.DeleteDish(ctx, DeleteDishRequest{UserID: claim.UserID, ID: id}); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to delete dish")
		errs.WriteGinError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
