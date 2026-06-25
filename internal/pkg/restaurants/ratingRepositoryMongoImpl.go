package restaurants

import (
	"context"
	"fmt"
	"time"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// RatingRepositoryMongoImpl implements restaurants.RatingRepository using the
// MongoDB client directly for the "ratings" collection. It combines rating
// submission and read operations.
type RatingRepositoryMongoImpl struct {
	client *mongo.Client
}

var _ restaurants.RatingRepository = &RatingRepositoryMongoImpl{}

func NewRatingRepositoryMongoImpl(client *mongo.Client) *RatingRepositoryMongoImpl {
	return &RatingRepositoryMongoImpl{
		client: client,
	}
}

func (r *RatingRepositoryMongoImpl) collection() *mongo.Collection {
	return r.client.Database(databaseName).Collection("ratings")
}

// SubmitRating validates and stores a new rating for a dish.
func (r *RatingRepositoryMongoImpl) SubmitRating(ctx context.Context, request restaurants.SubmitRatingRequest) (*restaurants.SubmitRatingResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for SubmitRating: %w", err)
	}

	rating := restaurants.Rating{
		ID:        uuid.New().String(),
		DishID:    request.DishID,
		UserID:    request.UserID,
		Score:     request.Score,
		Review:    request.Review,
		CreatedAt: time.Now(),
	}

	if _, err := r.collection().InsertOne(ctx, rating); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to store rating")
		return nil, fmt.Errorf("SubmitRating failed: %w", err)
	}

	return &restaurants.SubmitRatingResponse{
		Rating: rating,
	}, nil
}

// ListRatings returns a paginated list of ratings for a given dish.
func (r *RatingRepositoryMongoImpl) ListRatings(ctx context.Context, request restaurants.ListRatingsRequest) (*restaurants.ListRatingsResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for ListRatings: %w", err)
	}

	filter := bson.M{"dish_id": request.DishID}

	total, err := r.collection().CountDocuments(ctx, filter)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to count ratings")
		return nil, fmt.Errorf("ListRatings failed: %w", err)
	}

	opts := options.Find().SetSkip(int64(request.Offset)).SetLimit(int64(request.Limit))
	cursor, err := r.collection().Find(ctx, filter, opts)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to list ratings from store")
		return nil, fmt.Errorf("ListRatings failed: %w", err)
	}
	defer cursor.Close(ctx)

	var ratings []restaurants.Rating
	if err := cursor.All(ctx, &ratings); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to decode ratings")
		return nil, fmt.Errorf("ListRatings failed: %w", err)
	}

	return &restaurants.ListRatingsResponse{
		Ratings: ratings,
		Total:   total,
	}, nil
}
