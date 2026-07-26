package restaurants

import (
	"encoding/json"
	"net/http"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RestaurantRegistrationRESTAdaptor exposes restaurant registration over a REST API.
type RestaurantRegistrationRESTAdaptor struct {
	registrar RestaurantRegistrationService
}

// NewRestaurantRegistrationRESTAdaptor returns a new RestaurantRegistrationRESTAdaptor.
func NewRestaurantRegistrationRESTAdaptor(registrar RestaurantRegistrationService) *RestaurantRegistrationRESTAdaptor {
	return &RestaurantRegistrationRESTAdaptor{registrar: registrar}
}

type RegisterRestaurantRESTRequest struct {
	Name  string `json:"name"`
	City  string `json:"city"`
	Image string `json:"image"`
}

type RegisterRestaurantRESTResponse struct {
	Restaurant Restaurant `json:"restaurant"`
}

func (a *RestaurantRegistrationRESTAdaptor) RegisterRestaurant(c *gin.Context) {
	ctx := c.Request.Context()
	claim, ok := authentication.LoginClaimFromGinContext(c)
	if !ok {
		log.Ctx(ctx).Warn().Msg("no login claim in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var request RegisterRestaurantRESTRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to decode register restaurant request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := a.registrar.RegisterRestaurant(ctx, RegisterRestaurantRequest{
		UserID: claim.UserID,
		Name:   request.Name,
		City:   request.City,
		Image:  request.Image,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to register restaurant")
		errs.WriteGinError(c, err)
		return
	}

	c.JSON(http.StatusCreated, RegisterRestaurantRESTResponse{Restaurant: resp.Restaurant})
}
