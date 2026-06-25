package restaurants

import (
	"fmt"
	"strings"
)

func (r *CreateDishRequest) Validate() error {
	var reasons []string

	if r.UserID == "" {
		reasons = append(reasons, "UserID is required")
	}

	if r.Name == "" {
		reasons = append(reasons, "Name is required")
	}

	if r.Description == "" {
		reasons = append(reasons, "Description is required")
	}

	if r.Price <= 0 {
		reasons = append(reasons, "Price must be greater than 0")
	}

	if r.RestaurantID == "" {
		reasons = append(reasons, "RestaurantID is required")
	}

	if len(reasons) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))
	}

	return nil
}

func (r *GetDishRequest) Validate() error {
	var reasons []string

	if r.ID == "" {
		reasons = append(reasons, "ID is required")
	}

	if len(reasons) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))
	}

	return nil
}

func (r *ListDishesRequest) Validate() error {
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

func (r *SearchDishesRequest) Validate() error {
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

func (r *UpdateDishRequest) Validate() error {
	var reasons []string

	if r.UserID == "" {
		reasons = append(reasons, "UserID is required")
	}

	if r.ID == "" {
		reasons = append(reasons, "ID is required")
	}

	if r.Name == "" {
		reasons = append(reasons, "Name is required")
	}

	if r.Description == "" {
		reasons = append(reasons, "Description is required")
	}

	if r.Price <= 0 {
		reasons = append(reasons, "Price must be greater than 0")
	}

	if len(reasons) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))
	}

	return nil
}

func (r *DeleteDishRequest) Validate() error {
	var reasons []string

	if r.UserID == "" {
		reasons = append(reasons, "UserID is required")
	}

	if r.ID == "" {
		reasons = append(reasons, "ID is required")
	}

	if len(reasons) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))
	}

	return nil
}
