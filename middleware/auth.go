package middleware

import (
	"event-horizon/utils"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

/*********** JWT MIDDLEWARE FUNCTION  *************************************************

1. JWTMiddleware - Middleware function to protect routes using JWT authentication

2. echojwt.Config - Configuration struct for JWT middleware

3. SigningKey - Secret key used to sign and verify JWT tokens

4. TokenLookup - Location to look for the JWT token (e.g., header, query, cookie)

5. ParseTokenFunc - Custom function to parse and validate the JWT token

6. jwt.ParseWithClaims - Parse the JWT token with custom claims struct

7. utils.GetJWTSecret - Utility function to retrieve the JWT secret key

8. utils.JWTClaims - Custom JWT claims struct defined in utils package

9. echojwt.WithConfig - Create middleware function with the specified configuration

 ***************************************************************************************/

// JWTMiddleware returns the JWT middleware configured with the secret
func JWTMiddleware() echo.MiddlewareFunc {
	config := echojwt.Config{
		SigningKey:  []byte(utils.GetJWTSecret()), //! Get secret from utils
		TokenLookup: "header:Authorization",       //! Look for token in Authorization header

		ParseTokenFunc: func(c echo.Context, auth string) (interface{}, error) { //! Parse token using your custom JWTClaims struct
			//? Remove "Bearer " prefix if present
			tokenString := strings.TrimPrefix(auth, "Bearer ")
			tokenString = strings.TrimSpace(tokenString)

			token, err := jwt.ParseWithClaims(tokenString, &utils.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(utils.GetJWTSecret()), nil
			})

			if err != nil {
				println("Parse error:", err.Error())
				return nil, err
			}

			return token, nil
		},
	}
	//! Return the middleware function
	return echojwt.WithConfig(config)
}
