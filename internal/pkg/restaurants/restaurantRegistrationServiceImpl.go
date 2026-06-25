package restaurants

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	pkgMongo "github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/mongo"
	pkgRestaurants "github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
	"github.com/rs/zerolog/log"
)

type RestaurantRegistrationServiceImpl struct {
	restaurantRepository pkgRestaurants.RestaurantRepository
	userRepository       users.UserRepository
	transactionManager   pkgMongo.TransactionManager
}

func NewRestaurantRegistrationServiceImpl(
	restaurantRepository pkgRestaurants.RestaurantRepository,
	userRepository users.UserRepository,
	transactionManager pkgMongo.TransactionManager,
) *RestaurantRegistrationServiceImpl {
	return &RestaurantRegistrationServiceImpl{
		restaurantRepository: restaurantRepository,
		userRepository:       userRepository,
		transactionManager:   transactionManager,
	}
}

func (s *RestaurantRegistrationServiceImpl) RegisterRestaurant(ctx context.Context, request pkgRestaurants.RegisterRestaurantRequest) (*pkgRestaurants.RegisterRestaurantResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for RegisterRestaurant: %w", err)
	}

	// Creating the restaurant and promoting the user must happen atomically, so
	// run them inside a single transaction.
	var result *pkgRestaurants.RegisterRestaurantResponse
	err := s.transactionManager.RunInTransaction(ctx, func(sessionCtx context.Context) error {
		response, err := s.registerRestaurant(sessionCtx, request)
		if err != nil {
			return err
		}
		result = response
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// registerRestaurant performs the registration work against the supplied context,
// which is the session-scoped context from the surrounding transaction.
func (s *RestaurantRegistrationServiceImpl) registerRestaurant(ctx context.Context, request pkgRestaurants.RegisterRestaurantRequest) (*pkgRestaurants.RegisterRestaurantResponse, error) {
	// Fetch the current user.
	userResp, err := s.userRepository.GetUserByID(ctx, users.GetUserByIDRequest{ID: request.UserID})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get user")
		return nil, fmt.Errorf("RegisterRestaurant failed: %w", err)
	}
	user := userResp.User

	hasRole := slices.Contains(user.Roles, users.RoleRestaurantOwner)

	// Check if a restaurant already exists for this owner.
	var existing *pkgRestaurants.Restaurant
	existingResp, err := s.restaurantRepository.GetMyRestaurant(ctx, pkgRestaurants.GetMyRestaurantRequest{OwnerID: request.UserID})
	switch {
	case err == nil:
		existing = &existingResp.Restaurant
	case errors.Is(err, errs.ErrNotFound):
		existing = nil
	default:
		log.Ctx(ctx).Error().Err(err).Msg("failed to check existing restaurants")
		return nil, fmt.Errorf("RegisterRestaurant failed: %w", err)
	}

	// Both restaurant and role exist — already fully registered.
	if existing != nil && hasRole {
		return nil, fmt.Errorf("user already has a registered restaurant: %w", errs.ErrConflict)
	}

	// Restaurant exists but role is missing — recover from a previous partial failure.
	if existing != nil && !hasRole {
		if err := s.addRestaurantOwnerRole(ctx, user); err != nil {
			return nil, err
		}
		return &pkgRestaurants.RegisterRestaurantResponse{
			Restaurant: *existing,
		}, nil
	}

	// Create the restaurant.
	createResp, err := s.restaurantRepository.CreateRestaurant(ctx, pkgRestaurants.CreateRestaurantRequest{
		OwnerID: request.UserID,
		Name:    request.Name,
		City:    request.City,
		Image:   request.Image,
	})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to store restaurant")
		return nil, fmt.Errorf("RegisterRestaurant failed: %w", err)
	}

	// Update the user's roles to include RestaurantOwner.
	if err := s.addRestaurantOwnerRole(ctx, user); err != nil {
		return nil, err
	}

	return &pkgRestaurants.RegisterRestaurantResponse{
		Restaurant: createResp.Restaurant,
	}, nil
}

func (s *RestaurantRegistrationServiceImpl) addRestaurantOwnerRole(ctx context.Context, user users.User) error {
	newRoles := append(user.Roles, users.RoleRestaurantOwner)
	if _, err := s.userRepository.UpdateUserRoles(ctx, users.UpdateUserRolesRequest{UserID: user.ID, Roles: newRoles}); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to update user roles")
		return fmt.Errorf("RegisterRestaurant failed: %w", err)
	}
	return nil
}
