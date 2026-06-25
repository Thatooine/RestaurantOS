package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/errs"
	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const databaseName = "shire_shack"

// UserRepositoryMongoImpl implements users.UserRepository using the MongoDB
// client directly for the "users" collection. It combines user creation and
// read operations.
type UserRepositoryMongoImpl struct {
	client *mongo.Client
}

var _ users.UserRepository = &UserRepositoryMongoImpl{}

func NewUserRepositoryMongoImpl(client *mongo.Client) *UserRepositoryMongoImpl {
	return &UserRepositoryMongoImpl{
		client: client,
	}
}

func (r *UserRepositoryMongoImpl) collection() *mongo.Collection {
	return r.client.Database(databaseName).Collection("users")
}

// CreateUser stores a new user with the Customer role.
func (r *UserRepositoryMongoImpl) CreateUser(ctx context.Context, request users.CreateUserRequest) (*users.CreateUserResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for CreateUser: %w", err)
	}

	user := users.User{
		ID:           users.NewID(),
		Name:         request.Name,
		Email:        request.Email,
		Roles:        []users.Role{users.RoleCustomer},
		PasswordHash: request.PasswordHash,
	}

	if _, err := r.collection().InsertOne(ctx, user); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to store user")
		return nil, fmt.Errorf("CreateUser failed: %w", err)
	}

	return &users.CreateUserResponse{
		User: user,
	}, nil
}

// GetUser fetches a single user by Email.
func (r *UserRepositoryMongoImpl) GetUser(ctx context.Context, request users.GetUserRequest) (*users.GetUserResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for GetUser: %w", err)
	}

	var user users.User
	err := r.collection().FindOne(ctx, bson.M{"email": request.Email}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("user not found: %w", errs.ErrNotFound)
	}
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get user from store")
		return nil, fmt.Errorf("GetUser failed: %w", err)
	}

	return &users.GetUserResponse{
		User: user,
	}, nil
}

// GetUserByID fetches a single user by ID.
func (r *UserRepositoryMongoImpl) GetUserByID(ctx context.Context, request users.GetUserByIDRequest) (*users.GetUserByIDResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for GetUserByID: %w", err)
	}

	var user users.User
	err := r.collection().FindOne(ctx, bson.M{"id": request.ID}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("user not found: %w", errs.ErrNotFound)
	}
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to get user from store")
		return nil, fmt.Errorf("GetUserByID failed: %w", err)
	}

	return &users.GetUserByIDResponse{
		User: user,
	}, nil
}

// ListUsers returns a paginated list of users.
func (r *UserRepositoryMongoImpl) ListUsers(ctx context.Context, request users.ListUsersRequest) (*users.ListUsersResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for ListUsers: %w", err)
	}

	userList, total, err := r.list(ctx, bson.M{}, request.Offset, request.Limit)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to list users from store")
		return nil, fmt.Errorf("ListUsers failed: %w", err)
	}

	return &users.ListUsersResponse{
		Users: userList,
		Total: total,
	}, nil
}

// SearchUsers performs a case-insensitive regex search across user name and email.
func (r *UserRepositoryMongoImpl) SearchUsers(ctx context.Context, request users.SearchUsersRequest) (*users.SearchUsersResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for SearchUsers: %w", err)
	}

	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": request.Query, "$options": "i"}},
			{"email": bson.M{"$regex": request.Query, "$options": "i"}},
		},
	}

	userList, total, err := r.list(ctx, filter, request.Offset, request.Limit)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to search users from store")
		return nil, fmt.Errorf("SearchUsers failed: %w", err)
	}

	return &users.SearchUsersResponse{
		Users: userList,
		Total: total,
	}, nil
}

// UpdateUserRoles replaces the roles of the user identified by UserID.
func (r *UserRepositoryMongoImpl) UpdateUserRoles(ctx context.Context, request users.UpdateUserRolesRequest) (*users.UpdateUserRolesResponse, error) {
	if err := request.Validate(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("request validation failed")
		return nil, fmt.Errorf("invalid request for UpdateUserRoles: %w", err)
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var user users.User
	err := r.collection().FindOneAndUpdate(ctx, bson.M{"id": request.UserID}, bson.M{"$set": bson.M{"roles": request.Roles}}, opts).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("user not found: %w", errs.ErrNotFound)
	}
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to update user roles")
		return nil, fmt.Errorf("UpdateUserRoles failed: %w", err)
	}

	return &users.UpdateUserRolesResponse{
		User: user,
	}, nil
}

// list runs a paginated query against the users collection.
func (r *UserRepositoryMongoImpl) list(ctx context.Context, filter bson.M, offset int, limit int) ([]users.User, int64, error) {
	total, err := r.collection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit))
	cursor, err := r.collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer cursor.Close(ctx)

	var userList []users.User
	if err := cursor.All(ctx, &userList); err != nil {
		return nil, 0, fmt.Errorf("failed to decode users: %w", err)
	}

	return userList, total, nil
}
