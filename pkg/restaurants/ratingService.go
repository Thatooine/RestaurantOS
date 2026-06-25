package restaurants

import "context"

// RatingService defines the exposed application service for submitting and
// listing dish ratings. It is the REST-facing port (see RatingServiceRESTAdaptor)
// and delegates to the internal RatingRepository. The request/response DTOs are
// shared with RatingRepository (declared in ratingRepository.go).
type RatingService interface {
	SubmitRating(ctx context.Context, request SubmitRatingRequest) (*SubmitRatingResponse, error)
	ListRatings(ctx context.Context, request ListRatingsRequest) (*ListRatingsResponse, error)
}
