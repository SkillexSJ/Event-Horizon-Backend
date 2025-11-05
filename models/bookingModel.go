package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Booking struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID        bson.ObjectID `bson:"user_id" json:"user_id" validate:"required"` //? AUTO
	EventID       bson.ObjectID `bson:"event_id" json:"event_id" validate:"required"`
	TicketType    string        `bson:"ticket_type" json:"ticket_type" validate:"required,oneof=VIP Regular Student"`
	TransactionID string        `bson:"transaction_id" json:"transaction_id"`
	Quantity      int           `bson:"quantity" json:"quantity" validate:"required,gt=0"`
	TotalPaid     float64       `bson:"total_paid" json:"total_paid"` //? AUTO
	Status        string        `bson:"status" json:"status"` //? AUTO
	BookedAt      time.Time     `bson:"booked_at" json:"booked_at"` //? AUTO
}

// BookingWithDetails includes populated related data for API responses
type BookingWithDetails struct {
	Booking Booking `json:"booking"`
	Event   *Event  `json:"event,omitempty"`
	User    *User   `json:"user,omitempty"`
}
