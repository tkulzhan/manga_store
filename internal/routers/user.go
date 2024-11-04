package routers

import (
	"manga_store/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

type UserRouter struct {
	UserHandler handlers.UserHandler
}

func NewUserRouter() UserRouter {
	return UserRouter{
		UserHandler: handlers.NewUserHandler(),
	}
}

func (r UserRouter) SetupRoutes(app *fiber.App) {
	userGroup := app.Group("/user")

	userGroup.Delete("/", r.UserHandler.DeleteUser)
	userGroup.Post("/restore/:id", r.UserHandler.RestoreUser)

	userGroup.Get("/preferences/:id", r.UserHandler.GetUserByPreferences)
	userGroup.Get("/similar_users/:id", r.UserHandler.GetUserBySimilarUsers)
}
