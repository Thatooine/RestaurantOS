package restaurants

import (
	"context"
	"fmt"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const databaseName = "restaurantos"

// DishRepositoryMongoImpl implements restaurants.DishRepository using the MongoDB
// client directly for the "dishes" collection. It performs pure persistence only —
// ownership/authorization checks live in the DishService.
type DishRepositoryMongoImpl struct {
	client *mongo.Client
}

var _ restaurants.DishRepository = &DishRepositoryMongoImpl{}

func NewDishRepositoryMongoImpl(client *mongo.Client) *DishRepositoryMongoImpl {
	return &DishRepositoryMongoImpl{
		client: client,
	}
}

func (d *DishRepositoryMongoImpl) collection() *mongo.Collection {
	return d.client.Database(databaseName).Collection("dishes")
}

// CreateDish validates the request and inserts a new dish.
func (d *DishRepositoryMongoImpl) CreateDish(ctx context.Context, request restaurants.CreateDishRequest) (*restaurants.CreateDishResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for CreateDish: %w", err)
	}

	dish := restaurants.Dish{
		ID:           uuid.New().String(),
		Name:         request.Name,
		Description:  request.Description,
		Price:        request.Price,
		RestaurantID: request.RestaurantID,
		Image:        request.Image,
	}

	if _, err := d.collection().InsertOne(ctx, dish); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to store dish")
		return nil, fmt.Errorf("CreateDish failed: %w", err)
	}

	return &restaurants.CreateDishResponse{
		Dish: dish,
	}, nil
}

// GetDish fetches a single dish by ID.
func (d *DishRepositoryMongoImpl) GetDish(ctx context.Context, request restaurants.GetDishRequest) (*restaurants.GetDishResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for GetDish: %w", err)
	}

	var dish restaurants.Dish
	if err := d.collection().FindOne(ctx, bson.M{"id": request.ID}).Decode(&dish); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get dish from store")
		return nil, fmt.Errorf("GetDish failed: %w", err)
	}

	return &restaurants.GetDishResponse{
		Dish: dish,
	}, nil
}

// ListDishes returns a paginated list of dishes, optionally filtered by restaurant.
func (d *DishRepositoryMongoImpl) ListDishes(ctx context.Context, request restaurants.ListDishesRequest) (*restaurants.ListDishesResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for ListDishes: %w", err)
	}

	// Only filter by restaurant if provided
	filter := bson.M{}
	if request.RestaurantID != "" {
		filter["restaurant_id"] = request.RestaurantID
	}

	dishes, total, err := d.list(ctx, filter, request.Offset, request.Limit)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to list dishes from store")
		return nil, fmt.Errorf("ListDishes failed: %w", err)
	}

	return &restaurants.ListDishesResponse{
		Dishes: dishes,
		Total:  total,
	}, nil
}

// SearchDishes performs a case-insensitive regex search across dish name and description.
func (d *DishRepositoryMongoImpl) SearchDishes(ctx context.Context, request restaurants.SearchDishesRequest) (*restaurants.SearchDishesResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for SearchDishes: %w", err)
	}

	// Match dishes where name or description contains the query (case-insensitive)
	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": request.Query, "$options": "i"}},
			{"description": bson.M{"$regex": request.Query, "$options": "i"}},
		},
	}

	dishes, total, err := d.list(ctx, filter, request.Offset, request.Limit)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to search dishes from store")
		return nil, fmt.Errorf("SearchDishes failed: %w", err)
	}

	return &restaurants.SearchDishesResponse{
		Dishes: dishes,
		Total:  total,
	}, nil
}

// UpdateDish validates the request and updates the dish by ID.
func (d *DishRepositoryMongoImpl) UpdateDish(ctx context.Context, request restaurants.UpdateDishRequest) (*restaurants.UpdateDishResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for UpdateDish: %w", err)
	}

	update := bson.M{
		"name":        request.Name,
		"description": request.Description,
		"price":       request.Price,
		"image":       request.Image,
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var dish restaurants.Dish
	if err := d.collection().FindOneAndUpdate(ctx, bson.M{"id": request.ID}, bson.M{"$set": update}, opts).Decode(&dish); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to update dish in store")
		return nil, fmt.Errorf("UpdateDish failed: %w", err)
	}

	return &restaurants.UpdateDishResponse{
		Dish: dish,
	}, nil
}

// DeleteDish validates the request and removes the dish by ID.
func (d *DishRepositoryMongoImpl) DeleteDish(ctx context.Context, request restaurants.DeleteDishRequest) error {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return fmt.Errorf("invalid request for DeleteDish: %w", err)
	}

	result, err := d.collection().DeleteOne(ctx, bson.M{"id": request.ID})
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to remove dish from store")
		return fmt.Errorf("DeleteDish failed: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("dish with id %s not found", request.ID)
	}

	return nil
}

// list runs a paginated query against the dishes collection.
func (d *DishRepositoryMongoImpl) list(ctx context.Context, filter bson.M, offset int, limit int) ([]restaurants.Dish, int64, error) {
	total, err := d.collection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count dishes: %w", err)
	}

	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit))
	cursor, err := d.collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query dishes: %w", err)
	}
	defer cursor.Close(ctx)

	var dishes []restaurants.Dish
	if err := cursor.All(ctx, &dishes); err != nil {
		return nil, 0, fmt.Errorf("failed to decode dishes: %w", err)
	}

	return dishes, total, nil
}
