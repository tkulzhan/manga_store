package routers

import (
	"manga_store/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

type MangaRouter struct {
	mangaHandler handlers.MangaHandler
}

func NewMangaRouter() MangaRouter {
	return MangaRouter{
		mangaHandler: handlers.NewMangaHandler(),
	}
}

func (r MangaRouter) SetupRoutes(app *fiber.App) {
	mangaGroup := app.Group("/manga")

	mangaGroup.Get("/", r.mangaHandler.GetNewestManga)
	mangaGroup.Post("/search", r.mangaHandler.SearchManga)
	mangaGroup.Get("/popular", r.mangaHandler.GetPopularManga)
	mangaGroup.Post("/purchase", r.mangaHandler.PurchaseManga)
	
	mangaGroup.Get("/:id", r.mangaHandler.GetMangaByID)
	mangaGroup.Post("/:id/rate", r.mangaHandler.RateManga)
	mangaGroup.Patch("/:id/rate", r.mangaHandler.UpdateMangaRating)
	mangaGroup.Delete("/:id/rate", r.mangaHandler.RemoveMangaRating)
}
