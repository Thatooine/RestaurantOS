package restaurants

import (
	"fmt"
	"strings"
)

func (r *CreateRestaurantRequest) Validate() error {
	var reasons []string

	if r.OwnerID == "" {
		reasons = append(reasons, "OwnerID is required")
	}

	if r.Name == "" {
		reasons = append(reasons, "Name is required")
	}

	if r.City == "" {
		reasons = append(reasons, "City is required")
	}

	if len(reasons) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))
	}

	return nil
}

func (r *GetRestaurantRequest) Validate() error {
	var reasons []string

	if r.ID == "" {
		reasons = append(reasons, "ID is required")
	}

	if len(reasons) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))
	}

	return nil
}

func (r *ListRestaurantsRequest) Validate() error {
	var reasons []string

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

func (r *SearchRestaurantsRequest) Validate() error {
	var reasons []string

	if r.Query == "" {
		reasons = append(reasons, "Query is required")
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
