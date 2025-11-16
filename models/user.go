package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role defines user roles
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// SubscriptionStatus defines the user's subscription status
type SubscriptionStatus string

const (
	StatusActive   SubscriptionStatus = "active"
	StatusInactive SubscriptionStatus = "inactive"
	StatusTrial    SubscriptionStatus = "trial"
)

// UserSubscription embeds subscription info in the user document
type UserSubscription struct {
	Status  SubscriptionStatus `bson:"status" json:"status"`
	PlanID  primitive.ObjectID `bson:"planId" json:"planId"` // reference to subscriptionPlans
	EndDate time.Time          `bson:"endDate" json:"endDate"`
}

// User defines the user model
type User struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Email        string               `bson:"email" json:"email"`
	Password     string               `bson:"password" json:"-"` // never expose password
	Role         Role                 `bson:"role" json:"role"`
	Subscription UserSubscription     `bson:"subscription" json:"subscription"`
	Favorites    []primitive.ObjectID `bson:"favorites,omitempty" json:"favorites,omitempty"` // references words
	CreatedAt    time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time            `bson:"updatedAt" json:"updatedAt"`
}
