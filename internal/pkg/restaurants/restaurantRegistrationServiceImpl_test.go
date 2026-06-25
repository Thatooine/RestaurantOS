package restaurants

import (
	"context"
	"errors"
	"fmt"
	"testing"

	usersImpl "github.com/bash/the-dancing-pony-v2-rnyfbr/internal/pkg/users"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	pkgRestaurants "github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
)

func newTestRestaurantRegistrationService(restaurantRepository pkgRestaurants.RestaurantRepository, userRepository users.UserRepository) *RestaurantRegistrationServiceImpl {
	return &RestaurantRegistrationServiceImpl{
		restaurantRepository: restaurantRepository,
		userRepository:       userRepository,
	}
}

func TestRegisterRestaurant_ValidationFails(t *testing.T) {
	svc := newTestRestaurantRegistrationService(nil, nil)

	// Missing name and city
	_, err := svc.RegisterRestaurant(context.Background(), pkgRestaurants.RegisterRestaurantRequest{
		UserID: "user-1",
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestRegisterRestaurant_UserNotFound(t *testing.T) {
	userRepository := &usersImpl.UserRepositoryMock{
		GetUserByIDFn: func(_ context.Context, _ users.GetUserByIDRequest) (*users.GetUserByIDResponse, error) {
			return nil, fmt.Errorf("user not found")
		},
	}

	svc := newTestRestaurantRegistrationService(nil, userRepository)

	_, err := svc.RegisterRestaurant(context.Background(), pkgRestaurants.RegisterRestaurantRequest{
		UserID: "user-1",
		Name:   "The Prancing Pony",
		City:   "Bree",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRegisterRestaurant_AlreadyFullyRegistered(t *testing.T) {
	userRepository := &usersImpl.UserRepositoryMock{
		GetUserByIDFn: func(_ context.Context, _ users.GetUserByIDRequest) (*users.GetUserByIDResponse, error) {
			return &users.GetUserByIDResponse{
				User: users.User{
					ID:    "user-1",
					Name:  "Aragorn",
					Email: "aragorn@gondor.com",
					Roles: []users.Role{users.RoleCustomer, users.RoleRestaurantOwner},
				},
			}, nil
		},
	}

	restaurantRepository := &RestaurantRepositoryMock{
		GetMyRestaurantFn: func(_ context.Context, _ pkgRestaurants.GetMyRestaurantRequest) (*pkgRestaurants.GetRestaurantResponse, error) {
			return &pkgRestaurants.GetRestaurantResponse{
				Restaurant: pkgRestaurants.Restaurant{ID: "rest-1", OwnerID: "user-1", Name: "Existing Restaurant", City: "Bree"},
			}, nil
		},
	}

	svc := newTestRestaurantRegistrationService(restaurantRepository, userRepository)

	_, err := svc.RegisterRestaurant(context.Background(), pkgRestaurants.RegisterRestaurantRequest{
		UserID: "user-1",
		Name:   "New Restaurant",
		City:   "Rohan",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errs.ErrConflict) {
		t.Errorf("expected ErrConflict, got: %v", err)
	}
}

func TestRegisterRestaurant_RecoveryFromPartialFailure(t *testing.T) {
	roleUpdated := false
	// Restaurant exists but user lacks the RestaurantOwner role
	userRepository := &usersImpl.UserRepositoryMock{
		GetUserByIDFn: func(_ context.Context, _ users.GetUserByIDRequest) (*users.GetUserByIDResponse, error) {
			return &users.GetUserByIDResponse{
				User: users.User{
					ID:    "user-1",
					Name:  "Aragorn",
					Email: "aragorn@gondor.com",
					Roles: []users.Role{users.RoleCustomer}, // Missing RestaurantOwner
				},
			}, nil
		},
		UpdateUserRolesFn: func(_ context.Context, request users.UpdateUserRolesRequest) (*users.UpdateUserRolesResponse, error) {
			if request.UserID != "user-1" {
				t.Fatalf("expected update for user-1, got %s", request.UserID)
			}
			hasOwnerRole := false
			for _, r := range request.Roles {
				if r == users.RoleRestaurantOwner {
					hasOwnerRole = true
				}
			}
			if !hasOwnerRole {
				t.Fatal("expected RestaurantOwner role in update")
			}
			roleUpdated = true
			return &users.UpdateUserRolesResponse{}, nil
		},
	}

	existingRestaurant := pkgRestaurants.Restaurant{
		ID: "rest-1", OwnerID: "user-1", Name: "The Prancing Pony", City: "Bree",
	}

	restaurantRepository := &RestaurantRepositoryMock{
		GetMyRestaurantFn: func(_ context.Context, _ pkgRestaurants.GetMyRestaurantRequest) (*pkgRestaurants.GetRestaurantResponse, error) {
			return &pkgRestaurants.GetRestaurantResponse{Restaurant: existingRestaurant}, nil
		},
		CreateRestaurantFn: func(_ context.Context, _ pkgRestaurants.CreateRestaurantRequest) (*pkgRestaurants.CreateRestaurantResponse, error) {
			t.Fatal("CreateRestaurant should not be called when a restaurant already exists")
			return nil, nil
		},
	}

	svc := newTestRestaurantRegistrationService(restaurantRepository, userRepository)

	resp, err := svc.RegisterRestaurant(context.Background(), pkgRestaurants.RegisterRestaurantRequest{
		UserID: "user-1",
		Name:   "Ignored Name",
		City:   "Ignored City",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return the existing restaurant, not create a new one
	if resp.Restaurant.ID != "rest-1" {
		t.Errorf("expected existing restaurant ID rest-1, got %s", resp.Restaurant.ID)
	}
	if resp.Restaurant.Name != "The Prancing Pony" {
		t.Errorf("expected existing restaurant name, got %s", resp.Restaurant.Name)
	}
	if !roleUpdated {
		t.Error("expected user role to be updated")
	}
}

func TestRegisterRestaurant_NewRegistration(t *testing.T) {
	var createRequest pkgRestaurants.CreateRestaurantRequest
	roleUpdated := false

	userRepository := &usersImpl.UserRepositoryMock{
		GetUserByIDFn: func(_ context.Context, _ users.GetUserByIDRequest) (*users.GetUserByIDResponse, error) {
			return &users.GetUserByIDResponse{
				User: users.User{
					ID:    "user-1",
					Name:  "Aragorn",
					Email: "aragorn@gondor.com",
					Roles: []users.Role{users.RoleCustomer},
				},
			}, nil
		},
		UpdateUserRolesFn: func(_ context.Context, request users.UpdateUserRolesRequest) (*users.UpdateUserRolesResponse, error) {
			if request.UserID != "user-1" {
				t.Fatalf("expected update for user-1, got %s", request.UserID)
			}
			roleUpdated = true
			return &users.UpdateUserRolesResponse{}, nil
		},
	}

	restaurantRepository := &RestaurantRepositoryMock{
		GetMyRestaurantFn: func(_ context.Context, _ pkgRestaurants.GetMyRestaurantRequest) (*pkgRestaurants.GetRestaurantResponse, error) {
			// No existing restaurant for this owner
			return nil, fmt.Errorf("none: %w", errs.ErrNotFound)
		},
		CreateRestaurantFn: func(_ context.Context, request pkgRestaurants.CreateRestaurantRequest) (*pkgRestaurants.CreateRestaurantResponse, error) {
			createRequest = request
			return &pkgRestaurants.CreateRestaurantResponse{
				Restaurant: pkgRestaurants.Restaurant{
					ID:      "rest-1",
					OwnerID: request.OwnerID,
					Name:    request.Name,
					City:    request.City,
					Image:   request.Image,
				},
			}, nil
		},
	}

	svc := newTestRestaurantRegistrationService(restaurantRepository, userRepository)

	resp, err := svc.RegisterRestaurant(context.Background(), pkgRestaurants.RegisterRestaurantRequest{
		UserID: "user-1",
		Name:   "The Green Dragon",
		City:   "Hobbiton",
		Image:  "dragon.jpg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the repository received the correct create request
	if createRequest.OwnerID != "user-1" {
		t.Errorf("expected create ownerID user-1, got %s", createRequest.OwnerID)
	}
	if createRequest.Name != "The Green Dragon" {
		t.Errorf("expected create name The Green Dragon, got %s", createRequest.Name)
	}
	if createRequest.City != "Hobbiton" {
		t.Errorf("expected create city Hobbiton, got %s", createRequest.City)
	}

	// Verify restaurant in response
	if resp.Restaurant.Name != "The Green Dragon" {
		t.Errorf("expected name The Green Dragon, got %s", resp.Restaurant.Name)
	}
	if resp.Restaurant.OwnerID != "user-1" {
		t.Errorf("expected ownerID user-1, got %s", resp.Restaurant.OwnerID)
	}
	if resp.Restaurant.ID == "" {
		t.Error("expected a restaurant ID")
	}

	// Verify user role was updated
	if !roleUpdated {
		t.Error("expected user role to be updated")
	}
}

func TestRegisterRestaurant_CreateFails(t *testing.T) {
	userRepository := &usersImpl.UserRepositoryMock{
		GetUserByIDFn: func(_ context.Context, _ users.GetUserByIDRequest) (*users.GetUserByIDResponse, error) {
			return &users.GetUserByIDResponse{
				User: users.User{ID: "user-1", Roles: []users.Role{users.RoleCustomer}},
			}, nil
		},
	}

	restaurantRepository := &RestaurantRepositoryMock{
		GetMyRestaurantFn: func(_ context.Context, _ pkgRestaurants.GetMyRestaurantRequest) (*pkgRestaurants.GetRestaurantResponse, error) {
			return nil, fmt.Errorf("none: %w", errs.ErrNotFound)
		},
		CreateRestaurantFn: func(_ context.Context, _ pkgRestaurants.CreateRestaurantRequest) (*pkgRestaurants.CreateRestaurantResponse, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	svc := newTestRestaurantRegistrationService(restaurantRepository, userRepository)

	_, err := svc.RegisterRestaurant(context.Background(), pkgRestaurants.RegisterRestaurantRequest{
		UserID: "user-1",
		Name:   "Test",
		City:   "Test",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRegisterRestaurant_RoleUpdateFails(t *testing.T) {
	userRepository := &usersImpl.UserRepositoryMock{
		GetUserByIDFn: func(_ context.Context, _ users.GetUserByIDRequest) (*users.GetUserByIDResponse, error) {
			return &users.GetUserByIDResponse{
				User: users.User{ID: "user-1", Roles: []users.Role{users.RoleCustomer}},
			}, nil
		},
		UpdateUserRolesFn: func(_ context.Context, _ users.UpdateUserRolesRequest) (*users.UpdateUserRolesResponse, error) {
			return nil, fmt.Errorf("update failed")
		},
	}

	restaurantRepository := &RestaurantRepositoryMock{
		GetMyRestaurantFn: func(_ context.Context, _ pkgRestaurants.GetMyRestaurantRequest) (*pkgRestaurants.GetRestaurantResponse, error) {
			return nil, fmt.Errorf("none: %w", errs.ErrNotFound)
		},
		CreateRestaurantFn: func(_ context.Context, request pkgRestaurants.CreateRestaurantRequest) (*pkgRestaurants.CreateRestaurantResponse, error) {
			return &pkgRestaurants.CreateRestaurantResponse{
				Restaurant: pkgRestaurants.Restaurant{ID: "rest-1", OwnerID: request.OwnerID, Name: request.Name, City: request.City},
			}, nil
		},
	}

	svc := newTestRestaurantRegistrationService(restaurantRepository, userRepository)

	_, err := svc.RegisterRestaurant(context.Background(), pkgRestaurants.RegisterRestaurantRequest{
		UserID: "user-1",
		Name:   "Test",
		City:   "Test",
	})
	if err == nil {
		t.Fatal("expected error when role update fails, got nil")
	}
}
