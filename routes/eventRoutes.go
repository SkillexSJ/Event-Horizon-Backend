package routes

import (
	"event-horizon/controllers"
	"event-horizon/middleware"

	"github.com/labstack/echo/v4"
)

/********************* EVENT ROUTES ********************

GET /events/all           - Get all events (public)
GET /events/:id           - Get event by ID (public)
POST /events/create       - Create a new event (protected)
PUT /events/:id           - Update an event (protected)
DELETE /events/:id        - Delete an event (protected)

*/

func SetupEventRoutes(grp *echo.Group, cntrlr *controllers.EventController) {

	//! Protected routes (require JWT authentication)
	grp.POST("/create", cntrlr.CreateEvent, middleware.JWTMiddleware())
	grp.PUT("/:id", cntrlr.UpdateEvent, middleware.JWTMiddleware())
	grp.DELETE("/:id", cntrlr.DeleteEvent, middleware.JWTMiddleware())

	//! Public routes (no authentication required)
	grp.GET("/all", cntrlr.GetAllEvents)
	grp.GET("/:id", cntrlr.GetEventByID)

}
