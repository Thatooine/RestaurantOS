package restaurants

import "context"

// RestaurantRepository defines the persistence port for restaurants. This is an
// internal port and is NOT exposed over REST — see RestaurantService for the
// exposed read operations and RestaurantRegistrationService for registration.
type RestaurantRepository interface {
	CreateRestaurant(ctx context.Context, request CreateRestaurantRequest) (*CreateRestaurantResponse, error)
	GetRestaurant(ctx context.Context, request GetRestaurantRequest) (*GetRestaurantResponse, error)
	GetMyRestaurant(ctx context.Context, request GetMyRestaurantRequest) (*GetRestaurantResponse, error)
	ListRestaurants(ctx context.Context, request ListRestaurantsRequest) (*ListRestaurantsResponse, error)
	SearchRestaurants(ctx context.Context, request SearchRestaurantsRequest) (*SearchRestaurantsResponse, error)
}

// CreateRestaurant

type CreateRestaurantRequest struct {
	OwnerID string
	Name    string
	City    string
	Image   string
}

type CreateRestaurantResponse struct {
	Restaurant Restaurant
}

// GetRestaurant

type GetRestaurantRequest struct {
	ID string
}

type GetRestaurantResponse struct {
	Restaurant Restaurant
}

// GetMyRestaurant

type GetMyRestaurantRequest struct {
	OwnerID string
}

// ListRestaurants

type ListRestaurantsRequest struct {
	Offset int
	Limit  int
}

type ListRestaurantsResponse struct {
	Restaurants []Restaurant
	Total       int64
}

// SearchRestaurants

type SearchRestaurantsRequest struct {
	Query  string
	Offset int
	Limit  int
}

type SearchRestaurantsResponse struct {
	Restaurants []Restaurant
	Total       int64
}
