// repository/subscription_repository.go
package repository

import (
	"context"
	"time"

	"USDT_BackEnd/db"
	"USDT_BackEnd/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SubscriptionRepository struct{}

func (r *SubscriptionRepository) Insert(ctx context.Context, sub models.Subscription) error {
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()

	_, err := db.Database.Collection("subscriptions").InsertOne(ctx, sub)
	return err
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := db.Database.Collection("subscriptions").
		DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Subscription, error) {
	var sub models.Subscription
	err := db.Database.Collection("subscriptions").
		FindOne(ctx, bson.M{"_id": id}).Decode(&sub)
	return &sub, err
}

func (r *SubscriptionRepository) GetAll(ctx context.Context) ([]models.Subscription, error) {
	cursor, err := db.Database.Collection("subscriptions").
		Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subs []models.Subscription
	err = cursor.All(ctx, &subs)
	return subs, err
}

func (r *SubscriptionRepository) GetPaginated(
	ctx context.Context,
	search string,
	page int,
	limit int,
) ([]models.Subscription, int64, error) {

	filter := bson.M{}
	if search != "" {
		filter["header"] = bson.M{
			"$regex":   search,
			"$options": "i",
		}
	}

	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"createdAt": -1})

	coll := db.Database.Collection("subscriptions")

	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var subs []models.Subscription
	err = cursor.All(ctx, &subs)

	return subs, total, err
}
