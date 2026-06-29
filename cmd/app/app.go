package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	authenticationImpl "github.com/bash/the-dancing-pony-v2-rnyfbr/internal/pkg/authentication"
	rateLimitingImpl "github.com/bash/the-dancing-pony-v2-rnyfbr/internal/pkg/rateLimiting"
	restaurantsImpl "github.com/bash/the-dancing-pony-v2-rnyfbr/internal/pkg/restaurants"
	usersImpl "github.com/bash/the-dancing-pony-v2-rnyfbr/internal/pkg/users"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/authentication"
	pkgMongo "github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/mongo"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/rateLimiting"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
	"github.com/go-jose/go-jose/v4"
	"github.com/redis/go-redis/v9"
)

type App struct {
	UserRepository                       users.UserRepository
	UserService                          users.UserService
	DishRepository                       restaurants.DishRepository
	DishService                          restaurants.DishService
	RatingRepository                     restaurants.RatingRepository
	RatingService                        restaurants.RatingService
	RestaurantRepository                 restaurants.RestaurantRepository
	RestaurantService                    restaurants.RestaurantService
	RestaurantRegistrationService        restaurants.RestaurantRegistrationService
	EmailAndPasswordAuthenticatorService authentication.EmailAndPasswordAuthenticatorService
	UserRegistrationService              users.UserRegistrationService
	AccessTokenCreatorService            authentication.AccessTokenCreator
	AccessTokenValidatorService          authentication.AccessTokenValidator
	RateLimiter                          rateLimiting.RedisTokenBucketRateLimiter
}

func NewApp(ctx context.Context, conf *Config, secureConf *SecureConfig) (*App, error) {
	mongoClient, err := pkgMongo.NewClient(ctx, secureConf.MongoURI)
	if err != nil {
		return nil, err
	}

	userRepository := usersImpl.NewUserRepositoryMongoImpl(mongoClient)
	userService := usersImpl.NewUserServiceImpl(userRepository)

	dishRepository := restaurantsImpl.NewDishRepositoryMongoImpl(mongoClient)
	restaurantRepository := restaurantsImpl.NewRestaurantRepositoryMongoImpl(mongoClient)
	ratingRepository := restaurantsImpl.NewRatingRepositoryMongoImpl(mongoClient)

	transactionManager := pkgMongo.NewTransactionManager(mongoClient)

	dishService := restaurantsImpl.NewDishServiceImpl(dishRepository, restaurantRepository, userRepository)
	restaurantService := restaurantsImpl.NewRestaurantServiceImpl(restaurantRepository)
	ratingService := restaurantsImpl.NewRatingServiceImpl(ratingRepository)
	restaurantRegistrar := restaurantsImpl.NewRestaurantRegistrationServiceImpl(restaurantRepository, userRepository, transactionManager)

	block, _ := pem.Decode([]byte(secureConf.JWTPrivateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode JWT private key PEM")
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT private key: %w", err)
	}

	privateKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("JWT private key is not RSA")
	}

	tokenSigner, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: privateKey},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create token signer: %w", err)
	}

	// redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: secureConf.RedisURI,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	rateLimiter := rateLimitingImpl.NewRedisRateLimiterImpl(redisClient)

	// authentication services
	accessTokenCreator := authenticationImpl.NewAccessTokenCreatorServiceImpl(tokenSigner)
	accessTokenValidator := authenticationImpl.NewAccessTokenValidatorImpl(&privateKey.PublicKey)
	registrationService := usersImpl.NewUserRegistrationServiceImpl(accessTokenCreator, userRepository)
	emailPasswordAuthenticator := authenticationImpl.NewEmailAndPasswordAuthenticatorService(accessTokenCreator, userRepository)

	return &App{
		UserRepository:                       userRepository,
		UserService:                          userService,
		DishRepository:                       dishRepository,
		DishService:                          dishService,
		RatingRepository:                     ratingRepository,
		RatingService:                        ratingService,
		RestaurantRepository:                 restaurantRepository,
		RestaurantService:                    restaurantService,
		RestaurantRegistrationService:        restaurantRegistrar,
		UserRegistrationService:              registrationService,
		EmailAndPasswordAuthenticatorService: emailPasswordAuthenticator,
		AccessTokenCreatorService:            accessTokenCreator,
		AccessTokenValidatorService:          accessTokenValidator,
		RateLimiter:                          rateLimiter,
	}, nil
}
