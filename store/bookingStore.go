package store

import (
	"context"
	"errors"
	"event-horizon/models"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

/******************** MONGODB FUNCTIONALITY FOR BOOKINGS COLLECTION ********************

1. BSON MAPPING FOR BOOKINGS COLLECTION
2. InsertOne
3. FindOne
4. Find
5. DeleteOne
6. DeleteMany
7. UpdateOne

 ****************************************************************************************/

/************************** I DID THE FOLLOWING THINGS IN THIS FILE: *******************************************

1. Created BookingStore struct to manage booking-related database operations.

2. Implemented NewBookingStore constructor to initialize BookingStore with MongoDB collections.

3. Developed CreateBooking method to add new bookings, ensuring ticket availability and updating event ticket quantities within a transaction.

4. Created GetBookingByID method to fetch a specific booking by its ID.

5. Added GetBookingsByUserID method to retrieve all bookings made by a specific user.

6. Developed GetBookingsByEventID method to fetch all bookings for a specific event.

7. Implemented CancelBooking method to delete a booking and restore ticket quantities within a transaction.

8. Created GetAllBookings method to retrieve all bookings (admin function).

9. Added DeleteBookingsByEventID method to delete all bookings associated with a specific event.

************************************************************************************************************/

type BookingStore struct {
	db                *mongo.Database
	bookingCollection *mongo.Collection
	eventCollection   *mongo.Collection
}

func NewBookingStore(db *mongo.Database) *BookingStore {
	return &BookingStore{
		db:                db,
		bookingCollection: db.Collection("Bookings"),
		eventCollection:   db.Collection("Events"),
	}
}

// CreateBooking creates a booking with transaction to ensure data consistency
func (s *BookingStore) CreateBooking(ctx context.Context, booking *models.Booking) error {
	//? Start a session for transaction
	session, err := s.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	//! CALLBACK FUNCTION FOR TRANSACTION
	callback := func(sessCtx context.Context) (interface{}, error) {
		//? Get the event and find the ticket type
		var event models.Event
		eventFilter := bson.M{"_id": booking.EventID}

		if err := s.eventCollection.FindOne(sessCtx, eventFilter).Decode(&event); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, errors.New("event not found")
			}
			return nil, err
		}

		//? 2. Find the matching ticket type in the event's tickets array
		var selectedTicket *models.TicketInfo
		var ticketIndex int
		for i, ticket := range event.Tickets {
			if ticket.Type == booking.TicketType {
				selectedTicket = &ticket
				ticketIndex = i
				break
			}
		}

		if selectedTicket == nil {
			return nil, errors.New("ticket type not found for this event")
		}

		//? 3. Check ticket availability
		if selectedTicket.AvailableQuantity < booking.Quantity {
			return nil, errors.New("not enough tickets available")
		}

		//? 4. Calculate total price
		booking.TotalPaid = selectedTicket.Price * float64(booking.Quantity)
		booking.BookedAt = time.Now()
		booking.Status = "confirmed"

		//? 5. Insert booking
		result, err := s.bookingCollection.InsertOne(sessCtx, booking)
		if err != nil {
			return nil, err
		}
		booking.ID = result.InsertedID.(bson.ObjectID)

		//? 6. Update event's ticket available quantity using positional operator
		newAvailableQuantity := selectedTicket.AvailableQuantity - booking.Quantity

		//? Using array position index to update specific ticket
		ticketFieldPath := "tickets." + fmt.Sprint(ticketIndex) + ".available_quantity"
		eventUpdate := bson.M{
			"$set": bson.M{
				ticketFieldPath: newAvailableQuantity,
			},
		}

		if _, err := s.eventCollection.UpdateOne(sessCtx, eventFilter, eventUpdate); err != nil {
			return nil, err
		}

		return booking, nil
	}

	//? Execute transaction
	_, err = session.WithTransaction(ctx, callback)
	return err
}

// GetBookingByID retrieves a single booking by its ID
func (s *BookingStore) GetBookingByID(ctx context.Context, bookingID string) (*models.Booking, error) {
	objID, err := bson.ObjectIDFromHex(bookingID)
	if err != nil {
		return nil, errors.New("invalid booking id")
	}

	var booking models.Booking
	err = s.bookingCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&booking)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("booking not found")
		}
		return nil, err
	}

	return &booking, nil
}

// GetBookingsByUserID retrieves all bookings made by a specific user
func (s *BookingStore) GetBookingsByUserID(ctx context.Context, userID bson.ObjectID) ([]models.Booking, error) {
	var bookings []models.Booking

	filter := bson.M{"user_id": userID}
	cursor, err := s.bookingCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	if bookings == nil {
		bookings = []models.Booking{}
	}

	return bookings, nil
}

// GetBookingsByEventID retrieves all bookings for a specific event
func (s *BookingStore) GetBookingsByEventID(ctx context.Context, eventID bson.ObjectID) ([]models.Booking, error) {
	var bookings []models.Booking

	filter := bson.M{"event_id": eventID}
	cursor, err := s.bookingCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	if bookings == nil {
		bookings = []models.Booking{}
	}

	return bookings, nil
}

// CancelBooking deletes a booking and restores ticket quantity
func (s *BookingStore) CancelBooking(ctx context.Context, bookingID bson.ObjectID) error {
	//? Start a session for transaction
	session, err := s.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Define transaction callback
	callback := func(sessCtx context.Context) (interface{}, error) {
		//? 1. Get the booking
		var booking models.Booking
		bookingFilter := bson.M{"_id": bookingID}

		if err := s.bookingCollection.FindOne(sessCtx, bookingFilter).Decode(&booking); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, errors.New("booking not found")
			}
			return nil, err
		}

		//? Check if already cancelled
		if booking.Status == "cancelled" {
			return nil, errors.New("booking already cancelled")
		}

		//? 2. Get the event and find the ticket type
		var event models.Event
		eventFilter := bson.M{"_id": booking.EventID}

		if err := s.eventCollection.FindOne(sessCtx, eventFilter).Decode(&event); err != nil {
			//! If event doesn't exist
			if errors.Is(err, mongo.ErrNoDocuments) {
				//! Delete  booking without restoring tickets
				if _, err := s.bookingCollection.DeleteOne(sessCtx, bookingFilter); err != nil {
					return nil, err
				}
				return nil, nil
			}
			return nil, err
		}

		//? 3. Find the matching ticket type and restore quantity
		var ticketIndex int
		var found bool
		for i, ticket := range event.Tickets {
			if ticket.Type == booking.TicketType {
				ticketIndex = i
				found = true
				break
			}
		}

		if found {
			// 4. Restore ticket quantity
			newAvailableQuantity := event.Tickets[ticketIndex].AvailableQuantity + booking.Quantity
			ticketFieldPath := "tickets." + fmt.Sprint(ticketIndex) + ".available_quantity"
			eventUpdate := bson.M{
				"$set": bson.M{
					ticketFieldPath: newAvailableQuantity,
				},
			}

			if _, err := s.eventCollection.UpdateOne(sessCtx, eventFilter, eventUpdate); err != nil {
				return nil, err
			}
		}

		//? 5. Delete the booking
		if _, err := s.bookingCollection.DeleteOne(sessCtx, bookingFilter); err != nil {
			return nil, err
		}

		return nil, nil
	}

	// Execute transaction
	_, err = session.WithTransaction(ctx, callback)
	return err
}

// GetAllBookings retrieves all bookings (admin function)
func (s *BookingStore) GetAllBookings(ctx context.Context) ([]models.Booking, error) {
	var bookings []models.Booking

	cursor, err := s.bookingCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	if bookings == nil {
		bookings = []models.Booking{}
	}

	return bookings, nil
}

// DeleteBookingsByEventID deletes all bookings associated with a specific event
func (s *BookingStore) DeleteBookingsByEventID(ctx context.Context, eventID bson.ObjectID) (int64, error) {
	filter := bson.M{"event_id": eventID}

	result, err := s.bookingCollection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}
