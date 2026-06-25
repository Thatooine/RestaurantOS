package users

import (
	"encoding/json"
	"net/http"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
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

func (a *UserRegistrationRESTAdaptor) RegisterWithEmailAndPassword(w http.ResponseWriter, r *http.Request) {
	var request RegisterWithEmailAndPasswordRESTRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to decode register request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp, err := a.registration.RegisterWithEmailAndPassword(r.Context(), RegisterWithEmailAndPasswordRequest{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to register with email and password")
		errs.WriteHTTPError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    resp.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterRESTResponse{
		UserID: resp.UserID,
		Email:  resp.Email,
	})
}
