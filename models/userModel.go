package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string        `bson:"name" json:"name" validate:"required"`
	Email     string        `bson:"email" json:"email" validate:"required,email"`
	Password  string        `bson:"password" json:"password,omitempty" validate:"required,min=6"`
	IsHost    bool          `bson:"is_host" json:"is_host"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}

// UserPublic is the user data returned in API responses (without password)
type UserPublic struct {
	ID        bson.ObjectID `json:"id"`
	FirstName string        `json:"first_name"`
	LastName  string        `json:"last_name"`
	Email     string        `json:"email"`
	IsHost    bool          `json:"is_host"`
	CreatedAt time.Time     `json:"created_at"`
}
