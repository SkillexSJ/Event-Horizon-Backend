package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

/*********** JWT FUNCTIONALITY AND METHODS

1. jwt.registeredClaims - Standard JWT claims like exp, iat, nbf

2. jwt.RegisteredClaims - Embedding standard claims in custom claims struct

3. jwt.NewNumericDate - Create NumericDate for exp, iat, nbf

4. jwt.NewWithClaims - Create new JWT token with custom claims

5. jwt.SigningMethodHS256 - HMAC SHA256 signing method

6. token.SignedString - Sign token with secret key

7. jwt.Parse - Parse and validate JWT token

8. jwt.MapClaims - Generic map for JWT claims

*/

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT token for a user
func GenerateJWT(userID, email, name string) (string, error) {

	//? secret from environment variable
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key" //! fallback
	}

	//! Create claims
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)), // Token expires in 30 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	//! Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//! Sign token with secret
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetJWTSecret returns the JWT secret from environment
func GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key" //! fallback
	}
	return secret
}

// GetUserIDFromToken extracts the user ID from the JWT token in the context
func GetUserIDFromToken(c echo.Context) (string, error) {
	user := c.Get("user")
	if user == nil {
		fmt.Println("User not found in context")
		return "", errors.New("user not found in context")
	}

	token, ok := user.(*jwt.Token)
	if !ok {
		fmt.Println("Invalid token format")
		return "", errors.New("invalid token format")
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		if claims.UserID == "" {
			fmt.Println("UserID is empty in token")
			return "", errors.New("user_id is empty in token")
		}
		return claims.UserID, nil
	}

	//? Fallback to MapClaims if needed
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Invalid claims format 2nd attempt")
		return "", errors.New("invalid claims format")
	}

	userIDInterface, exists := claims["user_id"] //? Extract user_id from claims

	if !exists {
		fmt.Println("user_id claim not found")
		return "", errors.New("user_id claim not found")
	}

	userID, ok := userIDInterface.(string)
	//? Handle case where user_id is not a string
	if !ok {
		//* Sometimes float64
		if userIDFloat, ok := userIDInterface.(float64); ok {
			userID = fmt.Sprintf("%.0f", userIDFloat)
		} else {
			return "", fmt.Errorf("user_id has unexpected type: %T", userIDInterface)
		}
	}

	//? Check if userID is empty
	if userID == "" {
		return "", errors.New("user_id is empty")
	}

	println("Extracted user the userID from token")

	return userID, nil
}

func GetUserEmailFromToken(c echo.Context) (string, error) {

	//? Get the Authorization header
	authHeader := c.Request().Header.Get("Authorization")

	//? Check if header is present
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	//! Remove "Bearer " prefix
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	//! Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		fmt.Println("Invalid token while extracting email")
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	//? extra check
	if !ok {
		return "", errors.New("invalid claims format")
	}

	email, ok := claims["email"].(string)

	//? Check if email claim exists and is a string
	if !ok {
		return "", errors.New("email not found in claims")
	}

	return email, nil
}
