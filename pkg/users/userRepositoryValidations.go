package users

import (
	"fmt"
	"strings"
)

func (r *CreateUserRequest) Validate() error {
	var reasons []string

	if r.Name == "" {
		reasons = append(reasons, "Name is required")
	}

	if r.Email == "" {
		reasons = append(reasons, "Email is required")
	}

	if r.PasswordHash == "" {
		reasons = append(reasons, "PasswordHash is required")
	}

	if len(reasons) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))
	}

	return nil
}

func (r *GetUserRequest) Validate() error {
	var reasons []string

	if r.Email == "" {
		reasons = append(reasons, "Email is required")
	}

	if len(reasons) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))
	}

	return nil
}

func (r *GetUserByIDRequest) Validate() error {
	var reasons []string

	if r.ID == "" {
		reasons = append(reasons, "ID is required")
	}

	if len(reasons) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))
	}

	return nil
}

func (r *UpdateUserRolesRequest) Validate() error {
	var reasons []string

	if r.UserID == "" {
		reasons = append(reasons, "UserID is required")
	}

	if len(reasons) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(reasons, "; "))
	}

	return nil
}

func (r *ListUsersRequest) Validate() error {
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

func (r *SearchUsersRequest) Validate() error {
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
