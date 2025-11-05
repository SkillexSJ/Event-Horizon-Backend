package controllers

import (
	"event-horizon/models"
	"event-horizon/store"
	"event-horizon/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UserController struct {
	store *store.UserStore
}

func NewUserController(s *store.UserStore) *UserController {
	return &UserController{
		store: s,
	}
}

// Register functions
func (cntrlr *UserController) Register(c echo.Context) error {
	user := new(models.User)

	// 1. Bind Request
	if err := c.Bind(user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
	}

	// 2. Calling the Store password hashing will be done
	ctx := c.Request().Context()
	if err := cntrlr.store.CreateUser(ctx, user); err != nil {
		println("DEBUG Controller: Error creating user -", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	//! IMPORTANT: Fetch the newly created user to get the generated ID
	createdUser, err := cntrlr.store.FindUserByEmail(ctx, user.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve user")
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(createdUser.ID.Hex(), createdUser.Email, createdUser.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
	}

	//! password will be removed from response
	createdUser.Password = ""

	//   Response
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user":    user,
		"token":   token,
	})
}

// Login functions
func (cntrlr *UserController) Login(c echo.Context) error {
	//? login structure
	type LoginRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	//? instance of login request
	loginReq := new(LoginRequest)

	// Bind Request with the context
	if err := c.Bind(loginReq); err != nil {
		println("DEBUG Controller: Error parsing login request -", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid request payload",
			"error":   err.Error(),
		})
	}

	//? Find user by email
	ctx := c.Request().Context()
	user, err := cntrlr.store.FindUserByEmail(ctx, loginReq.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]interface{}{
			"message": "Invalid email or password",
			"error":   err.Error(),
		})
	}

	// Verify password
	if err := cntrlr.store.VerifyPassword(user.Password, loginReq.Password); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]interface{}{
			"message": "Invalid email or password",
			"error":   err.Error(),
		})
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.Hex(), user.Email, user.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to generate token",
			"error":   err.Error(),
		})
	}

	// Remove password from response
	user.Password = ""

	//? Send HTTP Response with JWT token
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"user":    user,
		"token":   token,
	})
}
