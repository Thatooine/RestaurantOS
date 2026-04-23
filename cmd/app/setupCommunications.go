package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/logger"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/metrics"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/rateLimiting"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func setupCommunications(serviceProviders ServiceProviders) {
	port := 8080

	router := mux.NewRouter()
	router.Use(logger.Middleware)
	router.Use(metrics.Middleware)

	// prometheus scrape endpoint
	router.Handle("/metrics", metrics.Handler()).Methods(http.MethodGet)

	// health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods(http.MethodGet)

	// auth routes (unauthenticated)
	ipRateLimiter := rateLimiting.NewIpRateLimiterMiddleware(serviceProviders.RateLimiter, 5, time.Minute)
	emailPasswordAdaptor := authentication.NewEmailAndPasswordAuthenticatorRESTAdaptor(serviceProviders.EmailAndPasswordAuthenticatorService)
	router.Handle("/api/v1/auth/login", ipRateLimiter(http.HandlerFunc(emailPasswordAdaptor.Login))).Methods(http.MethodPost)

	firebaseAdaptor := authentication.NewFirebaseAuthenticatorRESTAdaptor(serviceProviders.FirebaseAuthenticatorService)
	router.HandleFunc("/api/v1/auth/firebase", firebaseAdaptor.Login).Methods(http.MethodPost)

	registrationAdaptor := users.NewUserRegistrationRESTAdaptor(serviceProviders.UserRegistrationService)
	router.HandleFunc("/api/v1/auth/register", registrationAdaptor.RegisterWithEmailAndPassword).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/auth/register/firebase", registrationAdaptor.RegisterWithFirebaseToken).Methods(http.MethodPost)

	// authenticated API subrouter
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(authentication.NewAuthMiddleware(serviceProviders.AccessTokenValidatorService))
	api.Use(rateLimiting.NewUserRateLimiterMiddleware(serviceProviders.RateLimiter, 20, time.Second))

	// user routes

	userReaderAdaptor := users.NewUserReaderServiceRESTAdaptor(serviceProviders.UserReaderService)
	api.HandleFunc("/users", userReaderAdaptor.ListUsers).Methods(http.MethodGet)
	api.HandleFunc("/users/search", userReaderAdaptor.SearchUsers).Methods(http.MethodGet)
	api.HandleFunc("/users/{email}", userReaderAdaptor.GetUser).Methods(http.MethodGet)

	// restaurant routes

	restaurantRegistrationAdaptor := restaurants.NewRestaurantRegistrationRESTAdaptor(serviceProviders.RestaurantRegistrationService)
	api.HandleFunc("/restaurants/register", restaurantRegistrationAdaptor.RegisterRestaurant).Methods(http.MethodPost)

	restaurantReaderAdaptor := restaurants.NewRestaurantReaderServiceRESTAdaptor(serviceProviders.RestaurantReaderService)
	api.HandleFunc("/restaurants", restaurantReaderAdaptor.ListRestaurants).Methods(http.MethodGet)
	api.HandleFunc("/restaurants/mine", restaurantReaderAdaptor.GetMyRestaurant).Methods(http.MethodGet)
	api.HandleFunc("/restaurants/search", restaurantReaderAdaptor.SearchRestaurants).Methods(http.MethodGet)
	api.HandleFunc("/restaurants/{id}", restaurantReaderAdaptor.GetRestaurant).Methods(http.MethodGet)

	// dish routes
	dishWriterAdaptor := restaurants.NewDishWriterServiceRESTAdaptor(serviceProviders.DishWriterService)
	api.HandleFunc("/dishes", dishWriterAdaptor.CreateDish).Methods(http.MethodPost)
	api.HandleFunc("/dishes/{id}", dishWriterAdaptor.UpdateDish).Methods(http.MethodPut)
	api.HandleFunc("/dishes/{id}", dishWriterAdaptor.DeleteDish).Methods(http.MethodDelete)

	dishReaderAdaptor := restaurants.NewDishReaderServiceRESTAdaptor(serviceProviders.DishReaderService)
	api.HandleFunc("/dishes", dishReaderAdaptor.ListDishes).Methods(http.MethodGet)
	api.HandleFunc("/dishes/search", dishReaderAdaptor.SearchDishes).Methods(http.MethodGet)
	api.HandleFunc("/dishes/{id}", dishReaderAdaptor.GetDish).Methods(http.MethodGet)

	// rating routes
	ratingSubmitterAdaptor := restaurants.NewRatingSubmitterServiceRESTAdaptor(serviceProviders.RatingSubmitterService)
	api.HandleFunc("/dishes/{id}/ratings", ratingSubmitterAdaptor.SubmitRating).Methods(http.MethodPost)

	ratingReaderAdaptor := restaurants.NewRatingReaderServiceRESTAdaptor(serviceProviders.RatingReaderService)
	api.HandleFunc("/dishes/{id}/ratings", ratingReaderAdaptor.ListRatings).Methods(http.MethodGet)

	// start http server
	log.Info().Msgf("Starting HTTP server on port %d", port)
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), router); err != nil {
			log.Fatal().Err(err).Msg("http server has stopped")
		}
	}()
}
