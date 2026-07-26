package authentication

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// EmailAndPasswordAuthenticatorRESTAdaptor exposes email/password authentication over a REST API.
type EmailAndPasswordAuthenticatorRESTAdaptor struct {
	authenticator EmailAndPasswordAuthenticatorService
}

// NewEmailAndPasswordAuthenticatorRESTAdaptor returns a new EmailAndPasswordAuthenticatorRESTAdaptor.
func NewEmailAndPasswordAuthenticatorRESTAdaptor(
	authenticator EmailAndPasswordAuthenticatorService,
) *EmailAndPasswordAuthenticatorRESTAdaptor {
	return &EmailAndPasswordAuthenticatorRESTAdaptor{
		authenticator: authenticator,
	}
}

// EmailAndPasswordLoginRESTRequest is the expected JSON body for email/password login.
type EmailAndPasswordLoginRESTRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// EmailAndPasswordLoginRESTResponse is the JSON response after a successful login.
type EmailAndPasswordLoginRESTResponse struct {
	Token  string `json:"token"`
	UserID string `json:"userID"`
	Email  string `json:"email"`
}

// Login handles POST requests to authenticate a user with email and password.
func (a *EmailAndPasswordAuthenticatorRESTAdaptor) Login(c *gin.Context) {
	ctx := c.Request.Context()
	var request EmailAndPasswordLoginRESTRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to decode login request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := a.authenticator.AuthenticateWithEmailAndPassword(ctx, EmailAndPasswordAuthRequest{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to authenticate with email and password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, EmailAndPasswordLoginRESTResponse{
		Token:  resp.Token,
		UserID: resp.UserID,
		Email:  resp.Email,
	})
}
