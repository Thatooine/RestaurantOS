package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bash/the-dancing-pony-v2-rnyfbr/pkg/restaurants"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const seedRestaurantID = "00000000-0000-0000-0000-000000000001"

var seedRestaurant = restaurants.Restaurant{
	ID:      seedRestaurantID,
	OwnerID: "00000000-0000-0000-0000-000000000000",
	Name:    "The Dancing Pony",
	City:    "Bree",
}

var seedDishes = []restaurants.Dish{
	{
		ID:           "00000000-0000-0000-0000-000000000101",
		RestaurantID: seedRestaurantID,
		Name:         "Lembas Bread",
		Description:  "Elven waybread wrapped in mallorn leaves. One bite is enough to fill the stomach of a grown man.",
		Price:        100,
		Image:        "https://images.unsplash.com/photo-1509440159596-0249088772ff?w=400&h=300&fit=crop",
	},
	{
		ID:           "00000000-0000-0000-0000-000000000102",
		RestaurantID: seedRestaurantID,
		Name:         "Shire Mushroom Stew",
		Description:  "A hearty stew made with the finest mushrooms from Farmer Maggot's fields. Seasoned with herbs from the Shire.",
		Price:        150,
		Image:        "https://images.unsplash.com/photo-1547592166-23ac45744acd?w=400&h=300&fit=crop",
	},
	{
		ID:           "00000000-0000-0000-0000-000000000103",
		RestaurantID: seedRestaurantID,
		Name:         "Second Breakfast Platter",
		Description:  "Eggs, bacon, sausages, toast, tomatoes, and nice crispy hash browns. Elevenses not included.",
		Price:        200,
		Image:        "https://images.unsplash.com/photo-1533089860892-a7c6f0a88666?w=400&h=300&fit=crop",
	},
}

func seedDishesAndRestaurant(ctx context.Context, client *mongo.Client) {
	db := client.Database("restaurantos")

	// Clear and seed restaurant
	restaurantCol := db.Collection("restaurants")
	if _, err := restaurantCol.DeleteMany(ctx, bson.M{}); err != nil {
		log.Fatalf("failed to clear restaurants collection: %v", err)
	}
	fmt.Println("cleared restaurants collection")

	if _, err := restaurantCol.InsertOne(ctx, seedRestaurant); err != nil {
		log.Fatalf("failed to insert seed restaurant: %v", err)
	}
	fmt.Printf("seed restaurant created: %s (%s)\n", seedRestaurant.Name, seedRestaurant.ID)

	// Clear and seed dishes
	dishCol := db.Collection("dishes")
	if _, err := dishCol.DeleteMany(ctx, bson.M{}); err != nil {
		log.Fatalf("failed to clear dishes collection: %v", err)
	}
	fmt.Println("cleared dishes collection")

	for _, dish := range seedDishes {
		if _, err := dishCol.InsertOne(ctx, dish); err != nil {
			log.Fatalf("failed to insert dish %q: %v", dish.Name, err)
		}
		fmt.Printf("dish created: %s (%s)\n", dish.Name, dish.ID)
	}
}
