package routes

import (
	"event-horizon/controllers"
	"event-horizon/middleware"

	"github.com/labstack/echo/v4"
)

/** *********************  CATEGORY ROUTES   ********************

GET /categories                     - Get all categories
GET /categories/with-events        - Get all categories with their events
GET /categories/:id                - Get category by ID
GET /categories/:id/events         - Get events by category ID
GET /categories/name/:name/events  - Get events by category name
POST /categories/create            - Create a new category (protected)
PUT /categories/:id                - Update a category (protected)
DELETE /categories/:id             - Delete a category (protected)

*****************************************************/

func CategoryRoutes(grp *echo.Group, cc *controllers.CategoryController) {

	grp.GET("", cc.GetAllCategories)
	grp.GET("/with-events", cc.GetAllCategoriesWithEvents)
	grp.GET("/:id", cc.GetCategoryByID)
	grp.GET("/:id/events", cc.GetCategoryWithEvents)
	grp.GET("/name/:name/events", cc.GetEventsByCategoryName)

	// Protected routes (require authentication)
	grp.POST("/create", cc.CreateCategory, middleware.JWTMiddleware())
	grp.PUT("/:id", cc.UpdateCategory, middleware.JWTMiddleware())
	grp.DELETE("/:id", cc.DeleteCategory, middleware.JWTMiddleware())
}
