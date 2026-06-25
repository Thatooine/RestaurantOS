package restaurants

import (
	"context"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
)

// RatingServiceImpl implements restaurants.RatingService by delegating to the
// internal restaurants.RatingRepository.
type RatingServiceImpl struct {
	repository restaurants.RatingRepository
}

var _ restaurants.RatingService = &RatingServiceImpl{}

func NewRatingServiceImpl(repository restaurants.RatingRepository) *RatingServiceImpl {
	return &RatingServiceImpl{
		repository: repository,
	}
}

func (s *RatingServiceImpl) SubmitRating(ctx context.Context, request restaurants.SubmitRatingRequest) (*restaurants.SubmitRatingResponse, error) {
	return s.repository.SubmitRating(ctx, request)
}

func (s *RatingServiceImpl) ListRatings(ctx context.Context, request restaurants.ListRatingsRequest) (*restaurants.ListRatingsResponse, error) {
	return s.repository.ListRatings(ctx, request)
}
