package handlers

import (
	"manga_store/internal/helpers"
	"manga_store/internal/models"
	"manga_store/internal/services"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	mangas, err := h.mangaService.GetNewestManga(10)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(mangas)
}

func (h MangaHandler) SearchManga(c *fiber.Ctx) error {
	var request models.SearchMangaRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	limit := request.Limit
	if limit <= 0 {
		limit = 10
	}

	mangas, err := h.mangaService.SearchManga(request.Query, request.Genres, request.Author, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve manga",
		})
	}

	return c.JSON(mangas)
}

func (h MangaHandler) GetMangaByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Manga ID is required",
		})
	}

	manga, err := h.mangaService.GetMangaByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve manga",
		})
	}

	if manga == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Manga not found",
		})
	}

	return c.JSON(manga)
}

func (h MangaHandler) PurchaseManga(c *fiber.Ctx) error {
	var request models.PurchaseRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	encUserId := c.Cookies("data")
	decUserId := helpers.Decrypt(encUserId)

	userId, err := primitive.ObjectIDFromHex(decUserId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user id"})
	}
	mangaId, err := primitive.ObjectIDFromHex(request.MangaID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid manga id"})
	}

	err = h.mangaService.PurchaseManga(userId, mangaId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Purchase successful"})
}

func (h MangaHandler) GetPopularManga(c *fiber.Ctx) error {
	mangas, err := h.mangaService.GetPopularManga()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get popular manga"})
	}
	return c.Status(fiber.StatusOK).JSON(mangas)
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
