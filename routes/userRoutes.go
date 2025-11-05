package routes

import (
	"event-horizon/controllers"

	"github.com/labstack/echo/v4"
)

func UserRoutes(e *echo.Group, controller *controllers.UserController) {
	//! USER ROUTES
	e.POST("/register", controller.Register)
	e.POST("/login", controller.Login)
}
