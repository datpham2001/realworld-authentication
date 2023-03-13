package middleware

import (
	"net/http"
	"realworld-authentication/helper"
	"strings"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the Authorization header value
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
		}

		// Parse the token from the header
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		claims, err := helper.ValidateToken(tokenString)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		c.Set("userId", claims.UserID)

		return next(c)
	}
}
