package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const databaseName = "restaurantos"

func main() {
	ctx := context.Background()

	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017/?directConnection=true"))
	if err != nil {
		log.Fatalf("failed to connect to mongo: %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("failed to ping mongo: %v", err)
	}

	createRootUser(ctx, client)
	seedDishesAndRestaurant(ctx, client)
	ensureIndexes(ctx, client)
}

func ensureIndexes(ctx context.Context, client *mongo.Client) {
	indexes := []struct {
		collection string
		field      string
		unique     bool
	}{
		{"users", "id", true},
		{"users", "email", true},
		{"restaurants", "id", true},
		{"restaurants", "ownerID", true},
		{"dishes", "id", true},
		{"dishes", "restaurant_id", false},
		{"ratings", "id", true},
		{"ratings", "dish_id", false},
	}

	for _, idx := range indexes {
		collection := client.Database(databaseName).Collection(idx.collection)
		model := mongo.IndexModel{
			Keys:    bson.M{idx.field: 1},
			Options: options.Index().SetUnique(idx.unique),
		}
		if _, err := collection.Indexes().CreateOne(ctx, model); err != nil {
			log.Fatalf("failed to create index on %s.%s: %v", idx.collection, idx.field, err)
		}
		fmt.Printf("index ensured: %s.%s (unique=%v)\n", idx.collection, idx.field, idx.unique)
	}
}
