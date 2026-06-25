package restaurants

import (
	"context"

	pkgRestaurants "github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
)

var _ pkgRestaurants.RatingRepository = &RatingRepositoryMock{}

// RatingRepositoryMock is a hand-written mock of restaurants.RatingRepository.
type RatingRepositoryMock struct {
	SubmitRatingFn func(ctx context.Context, request pkgRestaurants.SubmitRatingRequest) (*pkgRestaurants.SubmitRatingResponse, error)
	ListRatingsFn  func(ctx context.Context, request pkgRestaurants.ListRatingsRequest) (*pkgRestaurants.ListRatingsResponse, error)
}

func (m *RatingRepositoryMock) SubmitRating(ctx context.Context, request pkgRestaurants.SubmitRatingRequest) (*pkgRestaurants.SubmitRatingResponse, error) {
	return m.SubmitRatingFn(ctx, request)
}

func (m *RatingRepositoryMock) ListRatings(ctx context.Context, request pkgRestaurants.ListRatingsRequest) (*pkgRestaurants.ListRatingsResponse, error) {
	return m.ListRatingsFn(ctx, request)
}
