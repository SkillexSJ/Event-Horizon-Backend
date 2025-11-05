// package routes

// import (
// 	"event-horizon/controllers"
// 	"event-horizon/middleware"

// 	"github.com/labstack/echo/v4"
// )

// func TicketRoutes(e *echo.Group, ticketController *controllers.TicketController) {
// 	// Protected routes (require authentication)
// 	e.POST("", ticketController.CreateTicketsForEvent, middleware.JWTMiddleware())
// 	e.PUT("/:id/quantity", ticketController.UpdateTicketQuantity, middleware.JWTMiddleware())

// 	// Public routes
// 	e.GET("/event/:eventId", ticketController.GetTicketsByEventID)
// 	e.GET("/:id", ticketController.GetTicketByID)
// 	e.POST("/:id/check-availability", ticketController.CheckTicketAvailability)
// }

package routes