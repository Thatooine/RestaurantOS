package restaurants

import (
	"context"

	pkgRestaurants "github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
)

var _ pkgRestaurants.RestaurantRepository = &RestaurantRepositoryMock{}

// RestaurantRepositoryMock is a hand-written mock of restaurants.RestaurantRepository.
type RestaurantRepositoryMock struct {
	CreateRestaurantFn  func(ctx context.Context, request pkgRestaurants.CreateRestaurantRequest) (*pkgRestaurants.CreateRestaurantResponse, error)
	GetRestaurantFn     func(ctx context.Context, request pkgRestaurants.GetRestaurantRequest) (*pkgRestaurants.GetRestaurantResponse, error)
	GetMyRestaurantFn   func(ctx context.Context, request pkgRestaurants.GetMyRestaurantRequest) (*pkgRestaurants.GetRestaurantResponse, error)
	ListRestaurantsFn   func(ctx context.Context, request pkgRestaurants.ListRestaurantsRequest) (*pkgRestaurants.ListRestaurantsResponse, error)
	SearchRestaurantsFn func(ctx context.Context, request pkgRestaurants.SearchRestaurantsRequest) (*pkgRestaurants.SearchRestaurantsResponse, error)
}

func (m *RestaurantRepositoryMock) CreateRestaurant(ctx context.Context, request pkgRestaurants.CreateRestaurantRequest) (*pkgRestaurants.CreateRestaurantResponse, error) {
	return m.CreateRestaurantFn(ctx, request)
}

func (m *RestaurantRepositoryMock) GetRestaurant(ctx context.Context, request pkgRestaurants.GetRestaurantRequest) (*pkgRestaurants.GetRestaurantResponse, error) {
	return m.GetRestaurantFn(ctx, request)
}

func (m *RestaurantRepositoryMock) GetMyRestaurant(ctx context.Context, request pkgRestaurants.GetMyRestaurantRequest) (*pkgRestaurants.GetRestaurantResponse, error) {
	return m.GetMyRestaurantFn(ctx, request)
}

func (m *RestaurantRepositoryMock) ListRestaurants(ctx context.Context, request pkgRestaurants.ListRestaurantsRequest) (*pkgRestaurants.ListRestaurantsResponse, error) {
	return m.ListRestaurantsFn(ctx, request)
}

func (m *RestaurantRepositoryMock) SearchRestaurants(ctx context.Context, request pkgRestaurants.SearchRestaurantsRequest) (*pkgRestaurants.SearchRestaurantsResponse, error) {
	return m.SearchRestaurantsFn(ctx, request)
}
