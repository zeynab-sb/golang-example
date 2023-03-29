package middleware

import (
	"github.com/labstack/echo/v4"
	"golang-example/utils"
	"net/http"
)

const (
	authorization      = "authorization"
	userIDContextField = "user_id"
)

func UserAuthorized() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			token := ctx.Request().Header.Get(authorization)
			id, err := utils.ValidateToken(token)
			if err != nil {
				return ctx.JSON(http.StatusUnauthorized, "Unauthorized")
			}

			ctx.Set(userIDContextField, id)

			return next(ctx)
		}
	}
}
