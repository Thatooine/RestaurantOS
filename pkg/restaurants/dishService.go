package restaurants

import "context"

// DishService defines the exposed application service for dishes. It is the
// REST-facing port (see DishServiceRESTAdaptor) and delegates persistence to the
// internal DishRepository while enforcing restaurant ownership on writes. The
// request/response DTOs are shared with DishRepository (declared in
// dishRepository.go).
type DishService interface {
	CreateDish(ctx context.Context, request CreateDishRequest) (*CreateDishResponse, error)
	GetDish(ctx context.Context, request GetDishRequest) (*GetDishResponse, error)
	ListDishes(ctx context.Context, request ListDishesRequest) (*ListDishesResponse, error)
	SearchDishes(ctx context.Context, request SearchDishesRequest) (*SearchDishesResponse, error)
	UpdateDish(ctx context.Context, request UpdateDishRequest) (*UpdateDishResponse, error)
	DeleteDish(ctx context.Context, request DeleteDishRequest) error
}
