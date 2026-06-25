package restaurants

import "context"

// DishRepository defines the persistence port for dishes. It combines read and
// write operations over the dishes store. This is an internal port and is NOT
// exposed over REST — see DishService for the exposed operations. The repository
// performs pure persistence only; ownership/authorization checks live in
// DishService.
type DishRepository interface {
	CreateDish(ctx context.Context, request CreateDishRequest) (*CreateDishResponse, error)
	GetDish(ctx context.Context, request GetDishRequest) (*GetDishResponse, error)
	ListDishes(ctx context.Context, request ListDishesRequest) (*ListDishesResponse, error)
	SearchDishes(ctx context.Context, request SearchDishesRequest) (*SearchDishesResponse, error)
	UpdateDish(ctx context.Context, request UpdateDishRequest) (*UpdateDishResponse, error)
	DeleteDish(ctx context.Context, request DeleteDishRequest) error
}

// CreateDish

type CreateDishRequest struct {
	UserID       string
	Name         string
	Description  string
	Price        float64
	RestaurantID string
	Image        string
}

type CreateDishResponse struct {
	Dish Dish
}

// GetDish

type GetDishRequest struct {
	ID string
}

type GetDishResponse struct {
	Dish Dish
}

// ListDishes

type ListDishesRequest struct {
	RestaurantID string
	Offset       int
	Limit        int
}

type ListDishesResponse struct {
	Dishes []Dish
	Total  int64
}

// SearchDishes

type SearchDishesRequest struct {
	Query  string
	Offset int
	Limit  int
}

type SearchDishesResponse struct {
	Dishes []Dish
	Total  int64
}

// UpdateDish

type UpdateDishRequest struct {
	UserID      string
	ID          string
	Name        string
	Description string
	Price       float64
	Image       string
}

type UpdateDishResponse struct {
	Dish Dish
}

// DeleteDish

type DeleteDishRequest struct {
	UserID string
	ID     string
}
