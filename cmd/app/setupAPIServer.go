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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

func setupAPIServer(app App) {
	port := 8080

	router := mux.NewRouter()
	router.Use(logger.Middleware)
	router.Use(metrics.Middleware)

	// prometheus scrape endpoint
	router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	// health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods(http.MethodGet)

	// auth routes (unauthenticated)
	ipRateLimiter := rateLimiting.NewIpRateLimiterMiddleware(app.RateLimiter, 5, time.Minute)
	emailPasswordAdaptor := authentication.NewEmailAndPasswordAuthenticatorRESTAdaptor(app.EmailAndPasswordAuthenticatorService)
	router.Handle("/api/v1/auth/login", ipRateLimiter(http.HandlerFunc(emailPasswordAdaptor.Login))).Methods(http.MethodPost)

	registrationAdaptor := users.NewUserRegistrationRESTAdaptor(app.UserRegistrationService)
	router.HandleFunc("/api/v1/auth/register", registrationAdaptor.RegisterWithEmailAndPassword).Methods(http.MethodPost)

	// authenticated API subrouter
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(authentication.NewAuthMiddleware(app.AccessTokenValidatorService))
	api.Use(rateLimiting.NewUserRateLimiterMiddleware(app.RateLimiter, 20, time.Second))

	// user routes

	userServiceAdaptor := users.NewUserServiceRESTAdaptor(app.UserService)
	api.HandleFunc("/users", userServiceAdaptor.ListUsers).Methods(http.MethodGet)
	api.HandleFunc("/users/search", userServiceAdaptor.SearchUsers).Methods(http.MethodGet)
	api.HandleFunc("/users/{email}", userServiceAdaptor.GetUser).Methods(http.MethodGet)

	// restaurant routes

	restaurantRegistrationAdaptor := restaurants.NewRestaurantRegistrationRESTAdaptor(app.RestaurantRegistrationService)
	api.HandleFunc("/restaurants/register", restaurantRegistrationAdaptor.RegisterRestaurant).Methods(http.MethodPost)

	restaurantServiceAdaptor := restaurants.NewRestaurantServiceRESTAdaptor(app.RestaurantService)
	api.HandleFunc("/restaurants", restaurantServiceAdaptor.ListRestaurants).Methods(http.MethodGet)
	api.HandleFunc("/restaurants/mine", restaurantServiceAdaptor.GetMyRestaurant).Methods(http.MethodGet)
	api.HandleFunc("/restaurants/search", restaurantServiceAdaptor.SearchRestaurants).Methods(http.MethodGet)
	api.HandleFunc("/restaurants/{id}", restaurantServiceAdaptor.GetRestaurant).Methods(http.MethodGet)

	// dish routes
	dishServiceAdaptor := restaurants.NewDishServiceRESTAdaptor(app.DishService)
	api.HandleFunc("/dishes", dishServiceAdaptor.CreateDish).Methods(http.MethodPost)
	api.HandleFunc("/dishes/{id}", dishServiceAdaptor.UpdateDish).Methods(http.MethodPut)
	api.HandleFunc("/dishes/{id}", dishServiceAdaptor.DeleteDish).Methods(http.MethodDelete)
	api.HandleFunc("/dishes", dishServiceAdaptor.ListDishes).Methods(http.MethodGet)
	api.HandleFunc("/dishes/search", dishServiceAdaptor.SearchDishes).Methods(http.MethodGet)
	api.HandleFunc("/dishes/{id}", dishServiceAdaptor.GetDish).Methods(http.MethodGet)

	// rating routes
	ratingServiceAdaptor := restaurants.NewRatingServiceRESTAdaptor(app.RatingService)
	api.HandleFunc("/dishes/{id}/ratings", ratingServiceAdaptor.SubmitRating).Methods(http.MethodPost)
	api.HandleFunc("/dishes/{id}/ratings", ratingServiceAdaptor.ListRatings).Methods(http.MethodGet)

	// start http server
	log.Info().Msgf("Starting HTTP server on port %d", port)
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), router); err != nil {
			log.Fatal().Err(err).Msg("http server has stopped")
		}
	}()
}
