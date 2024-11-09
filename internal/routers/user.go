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

	userGroup.Get("/", r.UserHandler.GetUser)
	userGroup.Delete("/", r.UserHandler.DeleteUser)
	userGroup.Post("/restore/:id", r.UserHandler.RestoreUser)

	userGroup.Get("/recs/preferences", r.UserHandler.GetRecsByPreferences)
	userGroup.Get("/recs/similar_users", r.UserHandler.GetRecsBySimilarUsers)
}
