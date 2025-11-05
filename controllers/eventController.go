package controllers

import (
	"event-horizon/models"
	"event-horizon/store"
	"event-horizon/utils"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

/******** ECHO FRAMEWORK FUNCTIONALITY ***********

1. echo.Context: Represents the context of the current HTTP request, providing methods to access request and response data.

2. c.Bind(): Binds the request body to a specified struct, useful for parsing JSON payloads.

3. c.JSON(): Sends a JSON response with a specified HTTP status code.

4. echo.NewHTTPError(): Creates a new HTTP error with a given status code and message.

5. c.Request().Context(): Retrieves the context from the current HTTP request, useful for passing request-scoped values and deadlines.


********************************************************/

//! THIS FILE HANDLES HTTP REQUESTS RELATED TO EVENTS AND SEND RESPONSES TO THE CLIENTS

/******************************* NOTE **************************************

I DID THE FOLLOWING THINGS IN THIS FILE:

1. Created EventController struct to manage "EVENT-RELATED" HTTP requests.

2. Implemented NewEventController constructor to CONNECT EventController with EventStore, CategoryStore, and UserStore references.

3. Developed CreateEvent method to handle event creation requests, including VALIDATION and AUTHENTICATION.

4. Created GetAllEvents method to retrieve and respond with all EVENTS.

5. Added GetEventByID method to fetch and respond with a specific EVENT by its ID.

6. Developed DeleteEvent method to handle event deletion requests, ensuring only HOSTS can delete their own events.

7. Created UpdateEvent method to handle event update requests, ensuring only HOSTS can update their own events.

********************************* NOTE ************************************/

// EventController manages HTTP requests related to events!
type EventController struct {
	eventStore    *store.EventStore
	categoryStore *store.CategoryStore
	userStore     *store.UserStore
}

// NewEventController creates a new EventController.
func NewEventController(eventStore *store.EventStore, categoryStore *store.CategoryStore, userStore *store.UserStore) *EventController {
	return &EventController{
		eventStore:    eventStore,
		categoryStore: categoryStore,
		userStore:     userStore,
	}
}

// ! CreateEvent handles the creation of a new event
func (cntrlr *EventController) CreateEvent(c echo.Context) error {
	event := new(models.Event)

	//? Bind Request (gets name, description, location, date, category_name from JSON)
	if err := c.Bind(event); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"message": "Cannot bind event data",
			"error":   err.Error(),
		})
	}

	//? Get user email from JWT token
	userEmail, err := utils.GetUserEmailFromToken(c)
	if err != nil {
		c.Logger().Error("TOKEN VALIDATION FAILED", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized - Invalid token")
	}

	ctx := c.Request().Context() //! CONTEXT FROM REQUEST

	//? Get user from database to get user ID
	user, err := cntrlr.userStore.FindUserByEmail(ctx, userEmail)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	//? Check if user is a host
	if !user.IsHost {
		return echo.NewHTTPError(http.StatusForbidden, "Only hosts can create events")
	}

	//? Set HostID from authenticated user
	event.HostID = user.ID

	//? Validate that category_name is provided
	if event.CategoryName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "category_name is required")
	}

	//? Validate that event date is not in the past (compare dates only, not time)
	today := time.Now().Truncate(24 * time.Hour)
	eventDate := event.Date.Truncate(24 * time.Hour)
	if eventDate.Before(today) {
		return echo.NewHTTPError(http.StatusBadRequest, "event date cannot be in the past")
	}

	//? Validate that end_time is after start_time
	if event.EndTime.Before(event.StartTime) || event.EndTime.Equal(event.StartTime) {
		return echo.NewHTTPError(http.StatusBadRequest, "end time must be after start time")
	}

	//? Create the event in database (CategoryID lookup happens in store)
	if err := cntrlr.eventStore.CreateEvent(ctx, event); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to create event",
			"error":   err.Error(),
		})
	}

	//? Convert to EventResponse and send HTTP Response
	eventResponse := &models.EventResponse{
		ID:           event.ID,
		Name:         event.Name,
		HostID:       event.HostID,
		CategoryName: event.CategoryName,
		Date:         event.Date,
		Location:     event.Location,
		Tickets:      event.Tickets,
	}

	return c.JSON(http.StatusCreated, eventResponse)
}

// ! GetAllEvents retrieves and returns all events
func (cntrlr *EventController) GetAllEvents(c echo.Context) error {
	ctx := c.Request().Context() //! CONTEXT FROM REQUEST

	//? Call the Store
	events, err := cntrlr.eventStore.GetAllEvents(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to retrieve events",
			"error":   err.Error(),
		})
	}

	//? Send HTTP Response
	return c.JSON(http.StatusOK, events)
}

// ! GetEventByID retrieves and returns a specific event by its ID
func (cntrlr *EventController) GetEventByID(c echo.Context) error {

	id := c.Param("id")          //! GET ID FROM URL PARAMS
	ctx := c.Request().Context() // ! CONTEXT FROM REQUEST

	//? Get the event from the database
	event, err := cntrlr.eventStore.GetEventByID(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{
			"message": "Event not found",
			"error":   err.Error(),
		})
	}

	//? Send HTTP Response
	return c.JSON(http.StatusOK, event)
}

// ! DeleteEvent deletes an event and all its associated bookings (HOST ONLY)
func (cntrlr *EventController) DeleteEvent(c echo.Context) error {
	id := c.Param("id")          //! GET ID FROM URL PARAMS
	ctx := c.Request().Context() //! CONTEXT FROM REQUEST

	//? Get user email from JWT token
	userEmail, err := utils.GetUserEmailFromToken(c)
	if err != nil {
		c.Logger().Error("TOKEN VALIDATION FAILED", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized - Invalid token")
	}

	//? Get user from database to get user ID
	user, err := cntrlr.userStore.FindUserByEmail(ctx, userEmail)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	//? Check if user is a HOST
	if !user.IsHost {
		return echo.NewHTTPError(http.StatusForbidden, "Only hosts can delete events")
	}

	//? Get the event to verify OWNERSHIP
	event, err := cntrlr.eventStore.GetEventByID(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "Event not found"})
	}

	//? Verify that the user is the HOST of this EVENT
	if event.HostID.Hex() != user.ID.Hex() {
		return echo.NewHTTPError(http.StatusForbidden, map[string]interface{}{
			"message": "You can only delete your own events",
			"error":   "forbidden",
		})
	}

	//? Delete the event (and CASCADE delete bookings)
	if err := cntrlr.eventStore.DeleteEvent(ctx, event.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to delete event",
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Event and all associated bookings deleted successfully",
	})
}

// ! UpdateEvent updates an event (host only)
func (cntrlr *EventController) UpdateEvent(c echo.Context) error {
	id := c.Param("id")          //! GET ID FROM URL PARAMS
	ctx := c.Request().Context() //! CONTEXT FROM REQUEST

	//? Get user email from JWT token
	userEmail, err := utils.GetUserEmailFromToken(c)
	if err != nil {
		c.Logger().Error("TOKEN VALIDATION FAILED", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized - Invalid token")
	}

	//? Get user from database to get user ID
	user, err := cntrlr.userStore.FindUserByEmail(ctx, userEmail)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	//? Check if user is a host
	if !user.IsHost {
		return echo.NewHTTPError(http.StatusForbidden, "Only hosts can update events")
	}

	//? Get the event to verify ownership
	existingEvent, err := cntrlr.eventStore.GetEventByID(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Event not found")
	}

	//? Verify that the user is the host of this event
	if existingEvent.HostID.Hex() != user.ID.Hex() {
		return echo.NewHTTPError(http.StatusForbidden, "You can only update your own events")
	}

	//? Bind the updated event data
	updatedEvent := new(models.Event)
	if err := c.Bind(updatedEvent); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot bind event data")
	}

	//? Preserve the original ID and HostID
	updatedEvent.ID = existingEvent.ID
	updatedEvent.HostID = existingEvent.HostID
	updatedEvent.CreatedAt = existingEvent.CreatedAt

	//? Validate that category_name is provided
	if updatedEvent.CategoryName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "category_name is required")
	}

	//? Validate that event date is not in the past (compare dates only, not time)
	today := time.Now().Truncate(24 * time.Hour)
	eventDate := updatedEvent.Date.Truncate(24 * time.Hour)
	if eventDate.Before(today) {
		return echo.NewHTTPError(http.StatusBadRequest, "event date cannot be in the past")
	}

	//? Validate that end_time is after start_time
	if updatedEvent.EndTime.Before(updatedEvent.StartTime) || updatedEvent.EndTime.Equal(updatedEvent.StartTime) {
		return echo.NewHTTPError(http.StatusBadRequest, "end time must be after start time")
	}

	//? Update the event in database
	if err := cntrlr.eventStore.UpdateEvent(ctx, updatedEvent); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update event: "+err.Error())
	}

	//? Convert to EventResponse and send HTTP Response
	eventResponse := &models.EventResponse{
		ID:           updatedEvent.ID,
		Name:         updatedEvent.Name,
		HostID:       updatedEvent.HostID,
		CategoryName: updatedEvent.CategoryName,
		Date:         updatedEvent.Date,
		Location:     updatedEvent.Location,
		Tickets:      updatedEvent.Tickets,
	}

	return c.JSON(http.StatusOK, eventResponse)
}
