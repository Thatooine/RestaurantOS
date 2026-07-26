package users

import (
	"encoding/json"
	"net/http"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// UserRegistrationRESTAdaptor exposes user registration over a REST API.
type UserRegistrationRESTAdaptor struct {
	registration UserRegistrationService
}

// NewUserRegistrationRESTAdaptor returns a new UserRegistrationRESTAdaptor.
func NewUserRegistrationRESTAdaptor(registration UserRegistrationService) *UserRegistrationRESTAdaptor {
	return &UserRegistrationRESTAdaptor{registration: registration}
}

// RegisterWithEmailAndPassword

type RegisterWithEmailAndPasswordRESTRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRESTResponse struct {
	UserID string `json:"userID"`
	Email  string `json:"email"`
}

func (a *UserRegistrationRESTAdaptor) RegisterWithEmailAndPassword(c *gin.Context) {
	ctx := c.Request.Context()
	var request RegisterWithEmailAndPasswordRESTRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to decode register request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := a.registration.RegisterWithEmailAndPassword(ctx, RegisterWithEmailAndPasswordRequest{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to register with email and password")
		errs.WriteGinError(c, err)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    resp.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	c.JSON(http.StatusCreated, RegisterRESTResponse{
		UserID: resp.UserID,
		Email:  resp.Email,
	})
}
