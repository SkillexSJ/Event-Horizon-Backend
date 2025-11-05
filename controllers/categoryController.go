package controllers

import (
	"event-horizon/models"
	"event-horizon/store"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/bson"
)

//! THIS FILE HANDLES HTTP REQUESTS RELATED TO CATEGORIES AND SEND RESPONSES TO THE CLIENTS

/******** ECHO FRAMEWORK FUNCTIONALITY ***********

1. echo.Context: Used to handle HTTP requests and responses.

2. c.Bind(&category): Binds incoming JSON request payload to the category struct.

3. c.JSON(statusCode, response): Sends a JSON response with the specified HTTP status code.

4. echo.NewHTTPError(statusCode, message): Creates a new HTTP error with a status code and message.

5. c.Param("id"): Retrieves path parameters from the URL.

6. c.Request().Context(): Gets the context of the current HTTP request for passing to store methods.

7. url.QueryUnescape(string): Decodes URL-encoded strings.

********************************************************/

/******************************* NOTE **************************************

I DID THE FOLLOWING THINGS IN THIS FILE:

1. Created CategoryController struct to manage category-related operations.

2. Implemented CreateCategory method to handle category creation requests.

3. Implemented GetAllCategories method to retrieve all categories.

4. Implemented GetAllCategoriesWithEvents method to fetch categories along with their associated events.

5. Implemented GetCategoryByID method to fetch a specific category by its ID.

6. Implemented GetCategoryWithEvents method to retrieve a category along with its events.

7. Implemented GetEventsByCategoryName method to fetch events based on category name.

8. Implemented UpdateCategory method to update category details.

9. Implemented DeleteCategory method to delete a category if it has no associated events.

10. Used echo.Context for handling HTTP requests and responses.

11. Validated request payloads and parameters.

12. Interacted with CategoryStore for data operations.

********************************* NOTE ************************************/

type CategoryController struct {
	categoryStore *store.CategoryStore
}

func NewCategoryController(categoryStore *store.CategoryStore) *CategoryController {
	return &CategoryController{
		categoryStore: categoryStore,
	}
}

// CreateCategory creates a new category
func (cc *CategoryController) CreateCategory(c echo.Context) error {
	var category models.Category

	if err := c.Bind(&category); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate
	if category.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Category name is required")
	}

	if err := cc.categoryStore.CreateCategory(c.Request().Context(), &category); err != nil {
		println("error creating categories FROM CATEGORY", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "cannot create category")
	}

	return c.JSON(http.StatusCreated, category)
}

// GetAllCategories retrieves all categories (simple list)
func (cc *CategoryController) GetAllCategories(c echo.Context) error {
	categories, err := cc.categoryStore.GetAllCategories(c.Request().Context())
	if err != nil {
		println("error getting categories FROM CATEGORY", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "cannot fetch categories")
	}

	return c.JSON(http.StatusOK, categories)
}

// GetAllCategoriesWithEvents retrieves all categories with their events
func (cc *CategoryController) GetAllCategoriesWithEvents(c echo.Context) error {
	categories, err := cc.categoryStore.GetAllCategoriesWithEvents(c.Request().Context())
	if err != nil {
		println("error getting categories with events FROM CATEGORY", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "cannot fetch categories with events")
	}

	return c.JSON(http.StatusOK, categories)
}

// GetCategoryByID retrieves a single category
func (cc *CategoryController) GetCategoryByID(c echo.Context) error {
	categoryID := c.Param("id") //! GET PARAM

	objID, err := bson.ObjectIDFromHex(categoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid category ID")
	}

	category, err := cc.categoryStore.GetCategoryByID(c.Request().Context(), objID)
	if err != nil {
		println("error getting category FROM CATEGORY", err.Error())
		return echo.NewHTTPError(http.StatusNotFound, "Category not found")
	}

	return c.JSON(http.StatusOK, category)
}

// GetCategoryWithEvents retrieves a category with all its events
func (cc *CategoryController) GetCategoryWithEvents(c echo.Context) error {
	categoryID := c.Param("id")

	objID, err := bson.ObjectIDFromHex(categoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid category ID")
	}

	categoryWithEvents, err := cc.categoryStore.GetCategoryWithEvents(c.Request().Context(), objID)
	if err != nil {
		println("error getting category with events FROM CATEGORY", err.Error())
		return echo.NewHTTPError(http.StatusNotFound, "Category not found")
	}

	return c.JSON(http.StatusOK, categoryWithEvents)
}

// GetEventsByCategoryName retrieves events by category name
func (cc *CategoryController) GetEventsByCategoryName(c echo.Context) error {
	categoryName := c.Param("name") // ! GET PARAM

	if categoryName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Category name is required")
	}

	// URL decode the category name
	decodedName, err := url.QueryUnescape(categoryName)
	if err != nil {
		println("DEBUG: Failed to decode category name:", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid category name")
	}

	// Get category by name first
	category, err := cc.categoryStore.GetCategoryByName(c.Request().Context(), decodedName)
	if err != nil {
		println("DEBUG: Category not found:", err.Error())
		return echo.NewHTTPError(http.StatusNotFound, "Category not found")
	}

	// Get events for this category
	categoryWithEvents, err := cc.categoryStore.GetCategoryWithEvents(c.Request().Context(), category.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "cannot fetch events for category")
	}

	return c.JSON(http.StatusOK, categoryWithEvents)
}

// UpdateCategory updates a category's details
func (cc *CategoryController) UpdateCategory(c echo.Context) error {
	categoryID := c.Param("id")

	objID, err := bson.ObjectIDFromHex(categoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid category ID")
	}

	var updates struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.Bind(&updates); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	//! Build update map
	updateMap := bson.M{}
	if updates.Name != "" {
		updateMap["name"] = updates.Name
	}
	if updates.Description != "" {
		updateMap["description"] = updates.Description
	}

	if len(updateMap) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "No fields to update")
	}

	if err := cc.categoryStore.UpdateCategory(c.Request().Context(), objID, updateMap); err != nil {
		println("error updating category FROM CATEGORY", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update category")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Category updated successfully"})
}

// DeleteCategory deletes a category and all its associated events and bookings (CASCADE)
func (cc *CategoryController) DeleteCategory(c echo.Context) error {
	categoryID := c.Param("id")

	objID, err := bson.ObjectIDFromHex(categoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid category ID")
	}

	//! cascade delete to remove category, events, and bookings
	if err := cc.categoryStore.DeleteCategoryWithCascade(c.Request().Context(), objID); err != nil {
		println("error deleting category FROM CATEGORY", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to delete category: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Category and all associated events deleted successfully",
	})
}
