package handlers

import (
	"manga_store/internal/services"

	"github.com/gofiber/fiber/v2"
)

type MangaHandler struct {
	mangaService services.MangaService
}

func NewMangaHandler() MangaHandler {
	return MangaHandler{
		mangaService: services.NewMangaService(),
	}
}

func (h MangaHandler) GetNewestManga(c *fiber.Ctx) error {
	return h.mangaService.GetNewestManga()
}

func (h MangaHandler) SearchManga(c *fiber.Ctx) error {
	return h.mangaService.SearchManga()
}

func (h MangaHandler) GetMangaByID(c *fiber.Ctx) error {
	return h.mangaService.GetMangaByID()
}

func (h MangaHandler) PurchaseManga(c *fiber.Ctx) error {
	return h.mangaService.PurchaseManga()
}

func (h MangaHandler) GetPopularManga(c *fiber.Ctx) error {
	return h.mangaService.GetPopularManga()
}

func (h MangaHandler) RateManga(c *fiber.Ctx) error {
	return h.mangaService.RateManga()
}

func (h MangaHandler) UpdateMangaRating(c *fiber.Ctx) error {
	return h.mangaService.UpdateMangaRating()
}

func (h MangaHandler) RemoveMangaRating(c *fiber.Ctx) error {
	return h.mangaService.RemoveMangaRating()
}
