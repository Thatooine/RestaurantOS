package main

import "testing"

func TestGetMongoURI(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		t.Setenv("MONGO_URI", "")
		if got := getMongoURI(); got != defaultMongoURI {
			t.Fatalf("URI: got %q, want %q", got, defaultMongoURI)
		}
	})

	t.Run("environment override", func(t *testing.T) {
		const customURI = "mongodb://localhost:27019/?replicaSet=rs0&directConnection=true"
		t.Setenv("MONGO_URI", customURI)
		if got := getMongoURI(); got != customURI {
			t.Fatalf("URI: got %q, want %q", got, customURI)
		}
	})
}
