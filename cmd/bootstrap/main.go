package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const databaseName = "restaurantos"
const defaultMongoURI = "mongodb://localhost:27017/?directConnection=true"

func main() {
	ctx := context.Background()

	client, err := mongo.Connect(options.Client().ApplyURI(getMongoURI()))
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

func getMongoURI() string {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		return defaultMongoURI
	}
	return mongoURI
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
