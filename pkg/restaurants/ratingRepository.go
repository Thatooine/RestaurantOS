package restaurants

import "context"

// RatingRepository defines the persistence port for ratings. It combines
// submission and read operations over the ratings store. This is an internal
// port and is NOT exposed over REST — see RatingService for the exposed
// operations.
type RatingRepository interface {
	SubmitRating(ctx context.Context, request SubmitRatingRequest) (*SubmitRatingResponse, error)
	ListRatings(ctx context.Context, request ListRatingsRequest) (*ListRatingsResponse, error)
}

// SubmitRating

type SubmitRatingRequest struct {
	DishID string
	UserID string
	Score  int
	Review string
}

type SubmitRatingResponse struct {
	Rating Rating
}

// ListRatings

type ListRatingsRequest struct {
	DishID string
	Offset int
	Limit  int
}

type ListRatingsResponse struct {
	Ratings []Rating
	Total   int64
}
