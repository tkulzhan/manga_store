package routers

import (
	"manga_store/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

type AuthRouter struct {
	authHandler handlers.AuthHandler
}

func NewAuthRouter() AuthRouter {
	return AuthRouter{
		authHandler: handlers.NewAuthHandler(),
	}
}

func (r AuthRouter) SetupRoutes(app *fiber.App) {
	authGroup := app.Group("auth")

	authGroup.Post("/register", r.authHandler.Register)
	authGroup.Post("/login", r.authHandler.Login)
	authGroup.Post("/logout", r.authHandler.Logout)
}