package main

import (
	"event-horizon/controllers"
	"event-horizon/db"
	"event-horizon/routes"
	"event-horizon/store"
	"event-horizon/utils"

	"net/http"
	"os"


	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type User struct {
	Name  string `json:"name" xml:"name"`
	Email string `json:"email" xml:"email"`
}

func main() {

	e := echo.New()
	database := db.ConnectDB()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"}, // frontend URLs
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true, //  using cookies or Authorization header
	}))

	// STARTING THE STORES
	userStore := store.NewUserStore(database)
	categoryStore := store.NewCategoryStore(database)
	eventStore := store.NewEventStore(database, categoryStore)
	bookingStore := store.NewBookingStore(database)

	// Set bookingStore reference in eventStore for cascade delete
	eventStore.SetBookingStore(bookingStore)
	
	// Set bookingStore reference in categoryStore for cascade delete
	categoryStore.SetBookingStore(bookingStore)

	// STARTING THE CONTROLLERS
	eventController := controllers.NewEventController(eventStore, categoryStore, userStore)
	userController := controllers.NewUserController(userStore)
	categoryController := controllers.NewCategoryController(categoryStore)
	bookingController := controllers.NewBookingController(bookingStore, eventStore)

	// START BACKGROUND SCHEDULER TO DELETE EXPIRED EVENTS
	utils.StartEventCleanupScheduler(eventStore)

	e.GET("/", func(c echo.Context) error {
		data := "Welcome to Event Horizon Backend!"
		return c.String(http.StatusOK, data)
	})

	// SETTING UP THE ROUTES
	eventGroup := e.Group("/api/events")
	userGroup := e.Group("/api/users")
	categoryGroup := e.Group("/api/categories")
	bookingGroup := e.Group("/api/bookings")

	routes.SetupEventRoutes(eventGroup, eventController)
	routes.UserRoutes(userGroup, userController)
	routes.CategoryRoutes(categoryGroup, categoryController)
	routes.SetupBookingRoutes(bookingGroup, bookingController)
	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
	
}
