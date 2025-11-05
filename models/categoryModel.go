package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Category struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string        `bson:"name" json:"name" validate:"required,min=2"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}

// CategoryWithEvents includes all events under this category
type CategoryWithEvents struct {
	Category   Category `json:"category"`
	Events     []Event  `json:"events"`
	EventCount int      `json:"event_count"`
}
