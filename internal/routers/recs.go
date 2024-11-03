package routers

import (
	"manga_store/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

type RecsRouter struct {
	RecsHandler handlers.RecsHandler
}

func NewRecsRouter() RecsRouter {
	return RecsRouter{
		RecsHandler: handlers.NewRecsHandler(),
	}
}

func (r RecsRouter) SetupRoutes(app *fiber.App) {
	recGroup := app.Group("/Recs")

	recGroup.Get("/preferences/:userId", r.RecsHandler.GetRecsByPreferences)
	recGroup.Get("/similar_users/:userId", r.RecsHandler.GetRecsBySimilarUsers)
}
