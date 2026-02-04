// models/subscription.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscription struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Header       string             `bson:"header" json:"header"`
	PlanID       primitive.ObjectID `bson:"planId" json:"planId"`
	SearchesLeft int                `bson:"searchesLeft" json:"searchesLeft"`
	Discount     float64            `bson:"discount" json:"discount"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}
