package restaurants

import (
	"fmt"
	"strings"
)

func (r *SubmitRatingRequest) Validate() error {
	var reasons []string

	if r.DishID == "" {
		reasons = append(reasons, "DishID is required")
	}

	if r.UserID == "" {
		reasons = append(reasons, "UserID is required")
	}

	if r.Score < 1 || r.Score > 5 {
		reasons = append(reasons, "Score must be between 1 and 5")
	}

	if len(reasons) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))
	}

	return nil
}

func (r *ListRatingsRequest) Validate() error {
	var reasons []string

	if r.DishID == "" {
		reasons = append(reasons, "DishID is required")
	}

	if r.Limit < 0 {
		reasons = append(reasons, "Limit must be >= 0")
	}

	if r.Offset < 0 {
		reasons = append(reasons, "Offset must be >= 0")
	}

	if len(reasons) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))
	}

	return nil
}
