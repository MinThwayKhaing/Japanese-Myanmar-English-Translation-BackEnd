package repository

import (
	"context"

	"USDT_BackEnd/db"
	"USDT_BackEnd/models"

	"go.mongodb.org/mongo-driver/bson"
)

type SubscriptionRepository struct{}

func (r *SubscriptionRepository) GetPrices(ctx context.Context) ([]models.SubscriptionPlan, error) {
	cursor, err := db.Database.Collection("subscriptionPlans").Find(ctx, bson.M{"isActive": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var plans []models.SubscriptionPlan
	if err := cursor.All(ctx, &plans); err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *SubscriptionRepository) UpdatePrices(ctx context.Context, monthly, yearly float64, freeMonths int) error {
	coll := db.Database.Collection("subscriptionPlans")
	_, err := coll.UpdateOne(ctx, bson.M{"type": "monthly"}, bson.M{"$set": bson.M{"price": monthly, "freeMonths": freeMonths}})
	if err != nil {
		return err
	}
	_, err = coll.UpdateOne(ctx, bson.M{"type": "yearly"}, bson.M{"$set": bson.M{"price": yearly, "freeMonths": freeMonths}})
	return err
}
