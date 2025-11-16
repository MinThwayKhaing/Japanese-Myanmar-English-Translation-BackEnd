package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubscriptionPlanType defines monthly/yearly
type SubscriptionPlanType string

const (
	MonthlyPlan SubscriptionPlanType = "monthly"
	YearlyPlan  SubscriptionPlanType = "yearly"
)

// SubscriptionPlan defines subscription plans in DB
type SubscriptionPlan struct {
	ID         primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Type       SubscriptionPlanType `bson:"type" json:"type"`
	Price      float64              `bson:"price" json:"price"` // float64 or Decimal128 if needed
	FreeMonths int                  `bson:"freeMonths" json:"freeMonths"`
	IsActive   bool                 `bson:"isActive" json:"isActive"` // only one active plan per type recommended
	CreatedAt  time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time            `bson:"updatedAt" json:"updatedAt"`
}
