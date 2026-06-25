package restaurants

import (
	"context"
	"fmt"
	"slices"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	pkgRestaurants "github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
	"github.com/rs/zerolog/log"
)

// DishServiceImpl implements restaurants.DishService. It enforces restaurant
// ownership on writes and delegates persistence to a restaurants.DishRepository.
// Restaurant lookups go through the RestaurantRepository and user lookups through
// the users.UserRepository.
type DishServiceImpl struct {
	repository           pkgRestaurants.DishRepository
	restaurantRepository pkgRestaurants.RestaurantRepository
	userRepository       users.UserRepository
}

var _ pkgRestaurants.DishService = &DishServiceImpl{}

func NewDishServiceImpl(
	repository pkgRestaurants.DishRepository,
	restaurantRepository pkgRestaurants.RestaurantRepository,
	userRepository users.UserRepository,
) *DishServiceImpl {
	return &DishServiceImpl{
		repository:           repository,
		restaurantRepository: restaurantRepository,
		userRepository:       userRepository,
	}
}

// verifyRestaurantOwnership checks that the user has the RestaurantOwner role
// and owns the restaurant identified by restaurantID.
func (s *DishServiceImpl) verifyRestaurantOwnership(ctx context.Context, userID string, restaurantID string) error {
	userResp, err := s.userRepository.GetUserByID(ctx, users.GetUserByIDRequest{ID: userID})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get user")
		return fmt.Errorf("failed to get user: %w", err)
	}

	if !slices.Contains(userResp.User.Roles, users.RoleRestaurantOwner) {
		return fmt.Errorf("user is not a restaurant owner: %w", errs.ErrForbidden)
	}

	restaurantResp, err := s.restaurantRepository.GetRestaurant(ctx, pkgRestaurants.GetRestaurantRequest{ID: restaurantID})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get restaurant")
		return fmt.Errorf("failed to get restaurant: %w", err)
	}

	if restaurantResp.Restaurant.OwnerID != userID {
		return fmt.Errorf("user does not own this restaurant: %w", errs.ErrForbidden)
	}

	return nil
}

// CreateDish verifies restaurant ownership then delegates to the repository.
func (s *DishServiceImpl) CreateDish(ctx context.Context, request pkgRestaurants.CreateDishRequest) (*pkgRestaurants.CreateDishResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for CreateDish: %w", err)
	}

	if err := s.verifyRestaurantOwnership(ctx, request.UserID, request.RestaurantID); err != nil {
		return nil, err
	}

	return s.repository.CreateDish(ctx, request)
}

// GetDish delegates to the repository.
func (s *DishServiceImpl) GetDish(ctx context.Context, request pkgRestaurants.GetDishRequest) (*pkgRestaurants.GetDishResponse, error) {
	return s.repository.GetDish(ctx, request)
}

// ListDishes delegates to the repository.
func (s *DishServiceImpl) ListDishes(ctx context.Context, request pkgRestaurants.ListDishesRequest) (*pkgRestaurants.ListDishesResponse, error) {
	return s.repository.ListDishes(ctx, request)
}

// SearchDishes delegates to the repository.
func (s *DishServiceImpl) SearchDishes(ctx context.Context, request pkgRestaurants.SearchDishesRequest) (*pkgRestaurants.SearchDishesResponse, error) {
	return s.repository.SearchDishes(ctx, request)
}

// UpdateDish looks up the dish, verifies ownership of its restaurant, then
// delegates the update to the repository.
func (s *DishServiceImpl) UpdateDish(ctx context.Context, request pkgRestaurants.UpdateDishRequest) (*pkgRestaurants.UpdateDishResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for UpdateDish: %w", err)
	}

	existing, err := s.repository.GetDish(ctx, pkgRestaurants.GetDishRequest{ID: request.ID})
	if err != nil {
		return nil, fmt.Errorf("UpdateDish failed: %w", err)
	}

	if err := s.verifyRestaurantOwnership(ctx, request.UserID, existing.Dish.RestaurantID); err != nil {
		return nil, err
	}

	return s.repository.UpdateDish(ctx, request)
}

// DeleteDish looks up the dish, verifies ownership of its restaurant, then
// delegates the delete to the repository.
func (s *DishServiceImpl) DeleteDish(ctx context.Context, request pkgRestaurants.DeleteDishRequest) error {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return fmt.Errorf("invalid request for DeleteDish: %w", err)
	}

	existing, err := s.repository.GetDish(ctx, pkgRestaurants.GetDishRequest{ID: request.ID})
	if err != nil {
		return fmt.Errorf("DeleteDish failed: %w", err)
	}

	if err := s.verifyRestaurantOwnership(ctx, request.UserID, existing.Dish.RestaurantID); err != nil {
		return err
	}

	return s.repository.DeleteDish(ctx, request)
}
