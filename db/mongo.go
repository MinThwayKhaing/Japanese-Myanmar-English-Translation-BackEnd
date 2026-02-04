package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"USDT_BackEnd/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


var Client *mongo.Client
var Database *mongo.Database

func ConnectDB(cfg *config.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("❌ Error connecting to MongoDB:", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("❌ Cannot ping MongoDB:", err)
	}

	Client = client
	Database = client.Database(cfg.Database)

	fmt.Println("✅ Connected to MongoDB!")

	// Auto setup
	ensureCollectionsAndIndexes(ctx)
	seedInitialData(ctx)
}

func ensureCollectionsAndIndexes(ctx context.Context) {
	collections := []string{"users", "words", "subscriptions"}

	existing, _ := Database.ListCollectionNames(ctx, bson.D{})
	existingMap := make(map[string]bool)
	for _, c := range existing {
		existingMap[c] = true
	}

	for _, name := range collections {
		if !existingMap[name] {
			if err := Database.CreateCollection(ctx, name); err != nil {
				log.Printf("⚠️ Could not create collection %s: %v", name, err)
			} else {
				log.Printf("🆕 Created collection: %s", name)
			}
		}
	}

	// ========== Create Indexes ==========

	// users: unique email
	userIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("unique_email"),
	}
	_, _ = Database.Collection("users").Indexes().CreateOne(ctx, userIdx)

	// words: text index for search
	wordIdx := mongo.IndexModel{
		Keys: bson.D{
			{Key: "subTerm", Value: "text"},
			{Key: "japanese", Value: "text"},
			{Key: "english", Value: "text"},
			{Key: "myanmar", Value: "text"},
		},
		Options: options.Index().SetName("text_search"),
	}
	_, _ = Database.Collection("words").Indexes().CreateOne(ctx, wordIdx)

	// subscriptions: ensure one document only
	subIdx := mongo.IndexModel{
			Keys:    bson.D{{Key: "header", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("unique_header"),
	}
	_, _ = Database.Collection("subscriptions").Indexes().CreateOne(ctx, subIdx)

	log.Println("✅ Collections and indexes verified/created.")
}
func seedInitialData(ctx context.Context) {
	subscriptions := Database.Collection("subscriptions")

	// 🔥 FORCE REMOVE OLD FIELDS (migration)
	_, err := subscriptions.UpdateMany(
		ctx,
		bson.M{}, // all docs
		bson.M{
			"$unset": bson.M{
				"type":         "",
				"monthlyPrice": "",
				"yearlyPrice":  "",
				"freeMonths":   "",
			},
			"$set": bson.M{
				"header":       "Default",
				"planId":       primitive.NilObjectID,
				"searchesLeft": 1000,
				"discount":     0,
				"updatedAt":    time.Now(),
			},
			"$setOnInsert": bson.M{
				"createdAt": time.Now(),
			},
		},
	)

	if err != nil {
		log.Println("❌ Subscription migration failed:", err)
		return
	}

	// 🔒 Ensure exactly ONE document exists
	count, _ := subscriptions.CountDocuments(ctx, bson.M{})
	if count == 0 {
		_, err := subscriptions.InsertOne(ctx, bson.M{
			"header":       "Default",
			"planId":       primitive.NilObjectID,
			"searchesLeft": 1000,
			"discount":     0,
			"createdAt":    time.Now(),
			"updatedAt":    time.Now(),
		})
		if err != nil {
			log.Println("❌ Insert default subscription failed:", err)
			return
		}
	}

	log.Println("🌱 Subscription schema migrated & ensured.")
}



