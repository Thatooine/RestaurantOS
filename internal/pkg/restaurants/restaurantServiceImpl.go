package restaurants

import (
	"context"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
)

// RestaurantServiceImpl implements restaurants.RestaurantService by delegating
// read operations to the internal restaurants.RestaurantRepository.
type RestaurantServiceImpl struct {
	repository restaurants.RestaurantRepository
}

var _ restaurants.RestaurantService = &RestaurantServiceImpl{}

func NewRestaurantServiceImpl(repository restaurants.RestaurantRepository) *RestaurantServiceImpl {
	return &RestaurantServiceImpl{
		repository: repository,
	}
}

func (s *RestaurantServiceImpl) GetRestaurant(ctx context.Context, request restaurants.GetRestaurantRequest) (*restaurants.GetRestaurantResponse, error) {
	return s.repository.GetRestaurant(ctx, request)
}

func (s *RestaurantServiceImpl) GetMyRestaurant(ctx context.Context, request restaurants.GetMyRestaurantRequest) (*restaurants.GetRestaurantResponse, error) {
	return s.repository.GetMyRestaurant(ctx, request)
}

func (s *RestaurantServiceImpl) ListRestaurants(ctx context.Context, request restaurants.ListRestaurantsRequest) (*restaurants.ListRestaurantsResponse, error) {
	return s.repository.ListRestaurants(ctx, request)
}

func (s *RestaurantServiceImpl) SearchRestaurants(ctx context.Context, request restaurants.SearchRestaurantsRequest) (*restaurants.SearchRestaurantsResponse, error) {
	return s.repository.SearchRestaurants(ctx, request)
}
