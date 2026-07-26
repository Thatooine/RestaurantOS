package main

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/logger"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/metrics"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/rateLimiting"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

func setupAPIServer(dependencies Dependencies) {
	port := 8080
	router := newAPIEngine(dependencies)

	log.Info().Msgf("Starting HTTP server on port %d", port)
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), router); err != nil {
			log.Fatal().Err(err).Msg("http server has stopped")
		}
	}()
}

func newAPIEngine(dependencies Dependencies) *gin.Engine {
	router := gin.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.HandleMethodNotAllowed = true
	router.RemoveExtraSlash = false

	router.Use(
		cleanPathMiddleware(),
		logger.Middleware(),
		metrics.Middleware(),
		gin.Recovery(),
	)

	router.NoRoute(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Data(http.StatusNotFound, "text/plain; charset=utf-8", []byte("404 page not found\n"))
	})
	router.NoMethod(func(c *gin.Context) {
		c.Writer.Header().Del("Allow")
		c.Status(http.StatusMethodNotAllowed)
		c.Writer.WriteHeaderNow()
	})

	// prometheus scrape endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// auth routes (unauthenticated)
	ipRateLimiter := rateLimiting.NewIpRateLimiterMiddleware(dependencies.RateLimiter, 5, time.Minute)
	emailPasswordAdaptor := authentication.NewEmailAndPasswordAuthenticatorRESTAdaptor(dependencies.EmailAndPasswordAuthenticatorService)
	router.POST("/api/v1/auth/login", ipRateLimiter, emailPasswordAdaptor.Login)

	registrationAdaptor := users.NewUserRegistrationRESTAdaptor(dependencies.UserRegistrationService)
	router.POST("/api/v1/auth/register", registrationAdaptor.RegisterWithEmailAndPassword)

	// authenticated API route group
	api := router.Group("/api/v1")
	api.Use(
		authentication.NewAuthMiddleware(dependencies.AccessTokenValidatorService),
		rateLimiting.NewUserRateLimiterMiddleware(dependencies.RateLimiter, 20, time.Second),
	)

	// user routes

	userServiceAdaptor := users.NewUserServiceRESTAdaptor(dependencies.UserService)
	api.GET("/users", userServiceAdaptor.ListUsers)
	api.GET("/users/search", userServiceAdaptor.SearchUsers)
	api.GET("/users/:email", userServiceAdaptor.GetUser)

	// restaurant routes

	restaurantRegistrationAdaptor := restaurants.NewRestaurantRegistrationRESTAdaptor(dependencies.RestaurantRegistrationService)
	api.POST("/restaurants/register", restaurantRegistrationAdaptor.RegisterRestaurant)

	restaurantServiceAdaptor := restaurants.NewRestaurantServiceRESTAdaptor(dependencies.RestaurantService)
	api.GET("/restaurants", restaurantServiceAdaptor.ListRestaurants)
	api.GET("/restaurants/mine", restaurantServiceAdaptor.GetMyRestaurant)
	api.GET("/restaurants/search", restaurantServiceAdaptor.SearchRestaurants)
	api.GET("/restaurants/:id", restaurantServiceAdaptor.GetRestaurant)

	// dish routes
	dishServiceAdaptor := restaurants.NewDishServiceRESTAdaptor(dependencies.DishService)
	api.POST("/dishes", dishServiceAdaptor.CreateDish)
	api.PUT("/dishes/:id", dishServiceAdaptor.UpdateDish)
	api.DELETE("/dishes/:id", dishServiceAdaptor.DeleteDish)
	api.GET("/dishes", dishServiceAdaptor.ListDishes)
	api.GET("/dishes/search", dishServiceAdaptor.SearchDishes)
	api.GET("/dishes/:id", dishServiceAdaptor.GetDish)

	// rating routes
	ratingServiceAdaptor := restaurants.NewRatingServiceRESTAdaptor(dependencies.RatingService)
	api.POST("/dishes/:id/ratings", ratingServiceAdaptor.SubmitRating)
	api.GET("/dishes/:id/ratings", ratingServiceAdaptor.ListRatings)

	return router
}

// cleanPathMiddleware preserves the API's existing path-cleaning behavior
// without enabling Gin's case-insensitive RedirectFixedPath behavior.
func cleanPathMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		cleanedPath := cleanPath(requestPath)
		if cleanedPath == requestPath {
			c.Next()
			return
		}

		redirectURL := *c.Request.URL
		redirectURL.Path = cleanedPath
		redirectURL.RawPath = ""
		c.Header("Location", redirectURL.String())
		c.Status(http.StatusMovedPermanently)
		c.Abort()
	}
}

func cleanPath(requestPath string) string {
	if requestPath == "" {
		return "/"
	}
	if requestPath[0] != '/' {
		requestPath = "/" + requestPath
	}

	cleanedPath := path.Clean(requestPath)
	if strings.HasSuffix(requestPath, "/") && cleanedPath != "/" {
		cleanedPath += "/"
	}
	return cleanedPath
}
