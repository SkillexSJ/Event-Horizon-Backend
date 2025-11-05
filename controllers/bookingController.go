package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"event-horizon/models"
	"event-horizon/store"
	"event-horizon/utils"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
)

//! THIS FILE HANDLES HTTP REQUESTS RELATED TO BOOKINGS AND SEND RESPONSES TO THE CLIENTS

/******** ECHO FRAMEWORK FUNCTIONALITY ***********

1. echo.Context: Used to handle HTTP requests and responses.

2. c.Bind(&bookingRequest): Binds incoming JSON request payload to the bookingRequest struct.

3. c.JSON(statusCode, response): Sends a JSON response with the specified HTTP status code.

4. echo.NewHTTPError(statusCode, message): Creates a new HTTP error with a status code and message.

5. c.Param("id"): Retrieves path parameters from the URL.

6. c.Request().Context(): Gets the context of the current HTTP request for passing to store methods.

********************************************************/

/******************************* NOTE **************************************

I DID THE FOLLOWING THINGS IN THIS FILE:

1. Created BookingController struct to manage booking-related operations.

2. Implemented CreateBooking method to handle booking creation requests:

3. Implemented GetUserBookings method to retrieve bookings for the authenticated user.

4. Implemented GetBookingByID method to fetch a specific booking by its ID.

5. Implemented CancelBooking method to allow users to cancel their bookings.

6. Used echo.Context for handling HTTP requests and responses.

7. Generated unique transaction IDs for each booking.

8. Validated request payloads and user authentication using JWT.

9. Interacted with BookingStore and EventStore for data operations.

10. Sent appropriate HTTP responses based on operation outcomes.

********************************* NOTE ************************************/

type BookingController struct {
	BookingStore *store.BookingStore
	EventStore   *store.EventStore
}

func NewBookingController(bookingStore *store.BookingStore, eventStore *store.EventStore) *BookingController {
	return &BookingController{
		BookingStore: bookingStore,
		EventStore:   eventStore,
	}
}

// generateTransactionID generates a random transaction ID
func generateTransactionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "TXN-" + hex.EncodeToString(bytes)
}

// CreateBooking handles booking creation
func (cntrlr *BookingController) CreateBooking(c echo.Context) error {

	//? REQUEST PAYLOAD STRUCT
	var bookingRequest struct {
		EventID    string `json:"event_id" validate:"required"`
		TicketType string `json:"ticket_type" validate:"required,oneof=VIP Regular Student"`
		Quantity   int    `json:"quantity" validate:"required,gt=0"`
	}

	//? Bind Request
	if err := c.Bind(&bookingRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload FROM BOOKING")
	}

	//? Validate quantity
	if bookingRequest.Quantity <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Quantity must be greater than zero FROM BOOKING")
	}

	//? Get user from JWT
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized FROM BOOKING")
	}

	//? Convert userID string to ObjectID
	userObjID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID FROM BOOKING")
	}

	//? Validate and convert event ID
	eventObjID, err := bson.ObjectIDFromHex(bookingRequest.EventID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid event ID FROM BOOKING")
	}

	//? Verify event exists
	event, err := cntrlr.EventStore.GetEventByID(c.Request().Context(), bookingRequest.EventID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Event not found FROM BOOKING")
	}

	//? Create booking object
	booking := models.Booking{
		UserID:        userObjID,
		EventID:       eventObjID,
		TicketType:    bookingRequest.TicketType,
		TransactionID: generateTransactionID(),
		Quantity:      bookingRequest.Quantity,
		Status:        "confirmed", // Auto-set
	}

	// Create booking (this handles ticket availability check and price calculation)
	if err := cntrlr.BookingStore.CreateBooking(c.Request().Context(), &booking); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error creating booking FROM BOOKING")
	}

	// Success Response
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message":        "Booking created successfully",
		"booking_id":     booking.ID.Hex(),
		"transaction_id": booking.TransactionID,
		"event_id":       event.ID.Hex(),
		"event_name":     event.Name,
		"ticket_type":    booking.TicketType,
		"quantity":       booking.Quantity,
		"total_paid":     booking.TotalPaid,
		"status":         booking.Status,
		"booked_at":      booking.BookedAt,
	})
}

// GetUserBookings retrieves all bookings for the authenticated user
func (cntrlr *BookingController) GetUserBookings(c echo.Context) error {
	//? Get user from JWT
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized FROM BOOKING")
	}

	//? Convert userID string to ObjectID
	userObjID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID FROM BOOKING")
	}

	bookings, err := cntrlr.BookingStore.GetBookingsByUserID(c.Request().Context(), userObjID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error retrieving bookings FROM BOOKING")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"bookings": bookings,
		"count":    len(bookings),
	})
}

// GetBookingByID retrieves a specific booking by ID
func (cntrlr *BookingController) GetBookingByID(c echo.Context) error {

	bookingID := c.Param("id") //! GET PARAM

	booking, err := cntrlr.BookingStore.GetBookingByID(c.Request().Context(), bookingID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Booking not found FROM BOOKING")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Booking retrieved successfully",
		"booking": booking,
	})
}

// CancelBooking deletes a booking and restores ticket quantity
func (cntrlr *BookingController) CancelBooking(c echo.Context) error {
	bookingID := c.Param("id") //! GET PARAM

	//? Convert to ObjectID
	bookingObjID, err := bson.ObjectIDFromHex(bookingID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid booking ID FROM FROM BOOKING")
	}

	//? Get user from JWT
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized FROM BOOKING")
	}

	//? Verify the booking belongs to the user
	booking, err := cntrlr.BookingStore.GetBookingByID(c.Request().Context(), bookingID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Booking not found FROM BOOKING")
	}

	if booking.UserID.Hex() != userID {
		return echo.NewHTTPError(http.StatusForbidden, "You can only cancel your own bookings FROM BOOKING")
	}

	//? Cancel (delete) the booking
	if err := cntrlr.BookingStore.CancelBooking(c.Request().Context(), bookingObjID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error cancelling booking FROM BOOKING")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Booking cancelled and deleted successfully",
	})
}

// GetAllBookings retrieves all bookings across all events (admin function)
func (cntrlr *BookingController) GetAllBookings(c echo.Context) error {
	bookings, err := cntrlr.BookingStore.GetAllBookings(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error retrieving all bookings FROM BOOKING")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"bookings": bookings,
		"count":    len(bookings),
	})
}
