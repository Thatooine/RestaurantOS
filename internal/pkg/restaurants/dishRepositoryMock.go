package restaurants

import (
	"context"

	pkgRestaurants "github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
)

var _ pkgRestaurants.DishRepository = &DishRepositoryMock{}

// DishRepositoryMock is a hand-written mock of restaurants.DishRepository.
type DishRepositoryMock struct {
	CreateDishFn   func(ctx context.Context, request pkgRestaurants.CreateDishRequest) (*pkgRestaurants.CreateDishResponse, error)
	GetDishFn      func(ctx context.Context, request pkgRestaurants.GetDishRequest) (*pkgRestaurants.GetDishResponse, error)
	ListDishesFn   func(ctx context.Context, request pkgRestaurants.ListDishesRequest) (*pkgRestaurants.ListDishesResponse, error)
	SearchDishesFn func(ctx context.Context, request pkgRestaurants.SearchDishesRequest) (*pkgRestaurants.SearchDishesResponse, error)
	UpdateDishFn   func(ctx context.Context, request pkgRestaurants.UpdateDishRequest) (*pkgRestaurants.UpdateDishResponse, error)
	DeleteDishFn   func(ctx context.Context, request pkgRestaurants.DeleteDishRequest) error
}

func (m *DishRepositoryMock) CreateDish(ctx context.Context, request pkgRestaurants.CreateDishRequest) (*pkgRestaurants.CreateDishResponse, error) {
	return m.CreateDishFn(ctx, request)
}

func (m *DishRepositoryMock) GetDish(ctx context.Context, request pkgRestaurants.GetDishRequest) (*pkgRestaurants.GetDishResponse, error) {
	return m.GetDishFn(ctx, request)
}

func (m *DishRepositoryMock) ListDishes(ctx context.Context, request pkgRestaurants.ListDishesRequest) (*pkgRestaurants.ListDishesResponse, error) {
	return m.ListDishesFn(ctx, request)
}

func (m *DishRepositoryMock) SearchDishes(ctx context.Context, request pkgRestaurants.SearchDishesRequest) (*pkgRestaurants.SearchDishesResponse, error) {
	return m.SearchDishesFn(ctx, request)
}

func (m *DishRepositoryMock) UpdateDish(ctx context.Context, request pkgRestaurants.UpdateDishRequest) (*pkgRestaurants.UpdateDishResponse, error) {
	return m.UpdateDishFn(ctx, request)
}

func (m *DishRepositoryMock) DeleteDish(ctx context.Context, request pkgRestaurants.DeleteDishRequest) error {
	return m.DeleteDishFn(ctx, request)
}
