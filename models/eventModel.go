package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TicketInfo struct {
	Type              string  `json:"type" bson:"type" validate:"required,oneof=VIP Regular Student"`
	Price             float64 `json:"price" bson:"price" validate:"required,gt=0"`
	TotalQuantity     int     `json:"total_quantity" bson:"total_quantity" validate:"required,gt=0"`
	AvailableQuantity int     `json:"available_quantity" bson:"available_quantity" validate:"required,gte=0"`
}

type Event struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	HostID       bson.ObjectID `bson:"host_id" json:"host_id" validate:"required"`
	CategoryName string        `bson:"category_name" json:"category_name" validate:"required"`
	Name         string        `bson:"name" json:"name" validate:"required"`
	Description  string        `bson:"description" json:"description"`
	Date         time.Time     `bson:"date" json:"date" validate:"required"`
	Location     string        `bson:"location" json:"location" validate:"required"`
	ImageURL     string        `bson:"image_url" json:"image_url"`
	StartTime    time.Time     `bson:"start_time" json:"start_time" validate:"required"`
	EndTime      time.Time     `bson:"end_time" json:"end_time" validate:"required"`
	CreatedAt    time.Time     `bson:"created_at" json:"created_at"`
	Tickets      []TicketInfo  `bson:"tickets" json:"tickets" validate:"dive,required"`
}

type EventResponse struct {
	ID           bson.ObjectID `json:"id,omitempty"`
	Name		 string        `json:"name"`
	HostID       bson.ObjectID `json:"host_id"`
	CategoryName string        `json:"category_name"`
	Date         time.Time     `json:"date"`
	Location     string        `json:"location"`
	Tickets      []TicketInfo  `json:"tickets"`
}







