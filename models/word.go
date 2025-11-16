package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Word struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SubTerm   string             `bson:"subTerm" json:"subTerm"`
	Japanese  string             `bson:"japanese" json:"japanese"`
	Myanmar   string             `bson:"myanmar" json:"myanmar"`
	English   string             `bson:"english" json:"english"`
	ImageURL  string             `bson:"imageUrl,omitempty" json:"imageUrl,omitempty"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
