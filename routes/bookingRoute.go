package routes

import (
	"event-horizon/controllers"
	"event-horizon/middleware"

	"github.com/labstack/echo/v4"
)

/** *********************  BOOKING ROUTES   ********************

POST /bookings/create         - Create a new booking (protected)
GET /bookings/user           - Get bookings for the authenticated user (protected)
GET /bookings/all            - Get all bookings (protected - admin)
GET /bookings/:id            - Get booking by ID (protected)
PUT /bookings/:id/cancel     - Cancel a booking (protected)

*****************************************************/

func SetupBookingRoutes(grp *echo.Group, cntrlr *controllers.BookingController) {
	grp.POST("/create", cntrlr.CreateBooking, middleware.JWTMiddleware())
	grp.GET("/user", cntrlr.GetUserBookings, middleware.JWTMiddleware())
	grp.GET("/all", cntrlr.GetAllBookings, middleware.JWTMiddleware())
	grp.GET("/:id", cntrlr.GetBookingByID, middleware.JWTMiddleware())
	grp.PUT("/:id/cancel", cntrlr.CancelBooking, middleware.JWTMiddleware())
}
