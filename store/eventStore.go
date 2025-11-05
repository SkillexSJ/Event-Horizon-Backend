package store

import (
	"context"
	"errors"
	"event-horizon/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

//! THIS FILE IS INTERNAL DATABASE CONNECTION FOR EVENTS COLLECTION IN MONGODB

/******************** MONGODB FUNCTIONALITY FOR EVENTS COLLECTION ********************

1. BSON MAPPING FOR EVENTS COLLECTION
2. InsertOne
3. FindOne
4. Find
5. DeleteOne
6. DeleteMany
7. UpdateOne

 ****************************************************************************************/

/************************** I DID THE FOLLOWING THINGS IN THIS FILE: *******************************************


1. Created EventStore struct to manage event-related database operations.

2. Implemented NewEventStore constructor to initialize EventStore with MongoDB collection and CategoryStore reference.

3. Added SetBookingStore method to set (BookingStore) reference for managing bookings related to events.

4. Developed CreateEvent method to add new events, ensuring category existence and unique event names.

5. Created toEventResponse helper function to convert Event model to EventResponse for API responses.

6. Implemented GetAllEvents method to retrieve all events from the database.

7. Added GetEventByID method to fetch a specific event by its ID.

8. Developed DeleteExpiredEvents method to remove events that have ended and their associated BOOKINGS.

9. Created DeleteEvent method to delete a specific event and its related bookings.

10. Implemented UpdateEvent method to modify existing event details, ensuring category validity.


************************************************************************************************************/

type EventStore struct {
	collection    *mongo.Collection
	categoryStore *CategoryStore
	bookingStore  *BookingStore
}

// NewEventStore !NewEventStore creates a new EventStore.
func NewEventStore(db *mongo.Database, categoryStore *CategoryStore) *EventStore {
	return &EventStore{
		collection:    db.Collection("Events"),
		categoryStore: categoryStore,
		bookingStore:  nil, // Will be set later via SetBookingStore
	}
}

// SetBookingStore ! SetBookingStore sets the bookingStore reference with events
func (s *EventStore) SetBookingStore(bookingStore *BookingStore) {
	s.bookingStore = bookingStore
}

// ! CREATE EVENT
func (s *EventStore) CreateEvent(ctx context.Context, event *models.Event) error {

	//? Validate that category exists by name
	_, err := s.categoryStore.GetCategoryByName(ctx, event.CategoryName)
	if err != nil {
		return errors.New("category not found: " + event.CategoryName)
	}

	//? Check for duplicate event name
	filter := bson.M{"name": event.Name}

	var existingEvent models.Event
	err = s.collection.FindOne(ctx, filter).Decode(&existingEvent)
	if err == nil {
		return errors.New("event with the same name already exists")
	}

	//? If the error is not ErrNoDocuments, return the error
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	//? 3. Set creation timestamp
	event.CreatedAt = time.Now()

	//? 4. Insert the event
	result, err := s.collection.InsertOne(ctx, event)
	if err != nil {
		return err
	}

	event.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// ! toEventResponse converts an Event to EventResponse
func toEventResponse(event *models.Event) *models.EventResponse {
	return &models.EventResponse{
		ID:           event.ID,
		Name:         event.Name,
		HostID:       event.HostID,
		CategoryName: event.CategoryName,
		Date:         event.Date,
		Location:     event.Location,
		Tickets:      event.Tickets,
	}
}

// ! GetAllEvents retrieves all events from the database and returns EventResponse
func (s *EventStore) GetAllEvents(ctx context.Context) ([]*models.Event, error) {
	var events []*models.Event

	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	if events == nil {
		events = []*models.Event{} //** Return empty slice
	}
	return events, nil
}

func (s *EventStore) GetEventByID(ctx context.Context, id string) (*models.Event, error) {
	var event models.Event

	bsonID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid event id")
	}

	filter := bson.M{"_id": bsonID}
	err = s.collection.FindOne(ctx, filter).Decode(&event)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	return &event, nil
}

// DeleteExpiredEvents deletes all events where end_time has passed and their associated bookings
func (s *EventStore) DeleteExpiredEvents(ctx context.Context) (int64, error) {

	//? Find events where end_time is before current time
	filter := bson.M{
		"end_time": bson.M{"$lt": time.Now()},
	}

	//? First, get all expired events to delete their bookings
	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return 0, err
	}
	//! ENSURE CURSOR IS CLOSED
	defer cursor.Close(ctx)

	var expiredEvents []models.Event
	if err = cursor.All(ctx, &expiredEvents); err != nil {
		return 0, err
	}

	//? Delete bookings for each expired event
	if s.bookingStore != nil {
		for _, event := range expiredEvents {
			_, err := s.bookingStore.DeleteBookingsByEventID(ctx, event.ID)
			if err != nil {
				return 0, errors.New("failed to delete bookings for expired event: " + err.Error())
			}
		}
	}

	//? Then delete the expired events
	result, err := s.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// DeleteEvent deletes an event by ID and all associated bookings
func (s *EventStore) DeleteEvent(ctx context.Context, id bson.ObjectID) error {
	// First, delete all bookings associated with this event
	if s.bookingStore != nil {
		_, err := s.bookingStore.DeleteBookingsByEventID(ctx, id)
		if err != nil {
			return errors.New("failed to delete associated bookings: " + err.Error())
		}
	}

	// Then delete the event
	filter := bson.M{"_id": id}

	result, err := s.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("event not found")
	}

	return nil
}

// UpdateEvent updates an event
func (s *EventStore) UpdateEvent(ctx context.Context, event *models.Event) error {
	//* Validate that category exists by name
	_, err := s.categoryStore.GetCategoryByName(ctx, event.CategoryName)
	if err != nil {
		return errors.New("category not found: " + event.CategoryName)
	}

	filter := bson.M{"_id": event.ID}
	update := bson.M{
		"$set": bson.M{
			"name":          event.Name,
			"category_name": event.CategoryName,
			"description":   event.Description,
			"date":          event.Date,
			"location":      event.Location,
			"image_url":     event.ImageURL,
			"start_time":    event.StartTime,
			"end_time":      event.EndTime,
			"tickets":       event.Tickets,
		},
	}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("event not found")
	}

	return nil
}
