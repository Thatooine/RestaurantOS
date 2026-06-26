package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/users"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

// defaultRootPassword is used when BOOTSTRAP_ROOT_PASSWORD is not set. It is for
// local development only.
const defaultRootPassword = "rootpassword"

func createRootUser(ctx context.Context, client *mongo.Client) {
	collection := client.Database("restaurantos").Collection("users")

	if _, err := collection.DeleteMany(ctx, bson.M{}); err != nil {
		log.Fatalf("failed to clear users collection: %v", err)
	}
	fmt.Println("cleared users collection")

	password := os.Getenv("BOOTSTRAP_ROOT_PASSWORD")
	if password == "" {
		password = defaultRootPassword
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash root user password: %v", err)
	}

	rootUser := users.User{
		ID:           "00000000-0000-0000-0000-000000000000",
		Name:         "Root User",
		Email:        "root+user@gmail.com",
		Roles:        []users.Role{users.RoleAdmin, users.RoleRestaurantOwner},
		PasswordHash: string(passwordHash),
	}

	if _, err := collection.InsertOne(ctx, rootUser); err != nil {
		log.Fatalf("failed to insert root user: %v", err)
	}

	fmt.Printf("root user created: %s (%s)\n", rootUser.Email, rootUser.ID)
}
