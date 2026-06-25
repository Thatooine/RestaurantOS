package restaurants

import (
	"context"
	"errors"
	"fmt"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// RestaurantRepositoryMongoImpl implements restaurants.RestaurantRepository using
// the MongoDB client directly for the "restaurants" collection.
type RestaurantRepositoryMongoImpl struct {
	client *mongo.Client
}

var _ restaurants.RestaurantRepository = &RestaurantRepositoryMongoImpl{}

func NewRestaurantRepositoryMongoImpl(client *mongo.Client) *RestaurantRepositoryMongoImpl {
	return &RestaurantRepositoryMongoImpl{
		client: client,
	}
}

func (r *RestaurantRepositoryMongoImpl) collection() *mongo.Collection {
	return r.client.Database(databaseName).Collection("restaurants")
}

// CreateRestaurant validates the request and inserts a new restaurant.
func (r *RestaurantRepositoryMongoImpl) CreateRestaurant(ctx context.Context, request restaurants.CreateRestaurantRequest) (*restaurants.CreateRestaurantResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for CreateRestaurant: %w", err)
	}

	restaurant := restaurants.Restaurant{
		ID:      uuid.New().String(),
		OwnerID: request.OwnerID,
		Name:    request.Name,
		City:    request.City,
		Image:   request.Image,
	}

	if _, err := r.collection().InsertOne(ctx, restaurant); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to store restaurant")
		return nil, fmt.Errorf("CreateRestaurant failed: %w", err)
	}

	return &restaurants.CreateRestaurantResponse{
		Restaurant: restaurant,
	}, nil
}

// GetRestaurant fetches a single restaurant by ID.
func (r *RestaurantRepositoryMongoImpl) GetRestaurant(ctx context.Context, request restaurants.GetRestaurantRequest) (*restaurants.GetRestaurantResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for GetRestaurant: %w", err)
	}

	var restaurant restaurants.Restaurant
	if err := r.collection().FindOne(ctx, bson.M{"id": request.ID}).Decode(&restaurant); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get restaurant from store")
		return nil, fmt.Errorf("GetRestaurant failed: %w", err)
	}

	return &restaurants.GetRestaurantResponse{
		Restaurant: restaurant,
	}, nil
}

// GetMyRestaurant fetches the restaurant owned by the given user.
func (r *RestaurantRepositoryMongoImpl) GetMyRestaurant(ctx context.Context, request restaurants.GetMyRestaurantRequest) (*restaurants.GetRestaurantResponse, error) {
	var restaurant restaurants.Restaurant
	err := r.collection().FindOne(ctx, bson.M{"ownerID": request.OwnerID}).Decode(&restaurant)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("no restaurant found for this user: %w", errs.ErrNotFound)
	}
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get restaurant from store")
		return nil, fmt.Errorf("GetMyRestaurant failed: %w", err)
	}

	return &restaurants.GetRestaurantResponse{
		Restaurant: restaurant,
	}, nil
}

// ListRestaurants returns a paginated list of restaurants.
func (r *RestaurantRepositoryMongoImpl) ListRestaurants(ctx context.Context, request restaurants.ListRestaurantsRequest) (*restaurants.ListRestaurantsResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for ListRestaurants: %w", err)
	}

	result, total, err := r.list(ctx, bson.M{}, request.Offset, request.Limit)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to list restaurants from store")
		return nil, fmt.Errorf("ListRestaurants failed: %w", err)
	}

	return &restaurants.ListRestaurantsResponse{
		Restaurants: result,
		Total:       total,
	}, nil
}

// SearchRestaurants performs a case-insensitive regex search across restaurant name and city.
func (r *RestaurantRepositoryMongoImpl) SearchRestaurants(ctx context.Context, request restaurants.SearchRestaurantsRequest) (*restaurants.SearchRestaurantsResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for SearchRestaurants: %w", err)
	}

	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": request.Query, "$options": "i"}},
			{"city": bson.M{"$regex": request.Query, "$options": "i"}},
		},
	}

	result, total, err := r.list(ctx, filter, request.Offset, request.Limit)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to search restaurants from store")
		return nil, fmt.Errorf("SearchRestaurants failed: %w", err)
	}

	return &restaurants.SearchRestaurantsResponse{
		Restaurants: result,
		Total:       total,
	}, nil
}

// list runs a paginated query against the restaurants collection.
func (r *RestaurantRepositoryMongoImpl) list(ctx context.Context, filter bson.M, offset int, limit int) ([]restaurants.Restaurant, int64, error) {
	total, err := r.collection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count restaurants: %w", err)
	}

	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit))
	cursor, err := r.collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query restaurants: %w", err)
	}
	defer cursor.Close(ctx)

	var result []restaurants.Restaurant
	if err := cursor.All(ctx, &result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode restaurants: %w", err)
	}

	return result, total, nil
}
