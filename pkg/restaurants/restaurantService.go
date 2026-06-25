package restaurants

import "context"

// RestaurantService defines the exposed application service for viewing, listing,
// and searching restaurants. It is the REST-facing port (see
// RestaurantServiceRESTAdaptor) and delegates to the internal
// RestaurantRepository. The request/response DTOs are shared with
// RestaurantRepository (declared in restaurantRepository.go).
type RestaurantService interface {
	GetRestaurant(ctx context.Context, request GetRestaurantRequest) (*GetRestaurantResponse, error)
	GetMyRestaurant(ctx context.Context, request GetMyRestaurantRequest) (*GetRestaurantResponse, error)
	ListRestaurants(ctx context.Context, request ListRestaurantsRequest) (*ListRestaurantsResponse, error)
	SearchRestaurants(ctx context.Context, request SearchRestaurantsRequest) (*SearchRestaurantsResponse, error)
}
