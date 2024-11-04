package handlers

import (
	"manga_store/internal/helpers"
	"manga_store/internal/models"
	"manga_store/internal/services"
	"strconv"

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

func (h MangaHandler) CreateManga(c *fiber.Ctx) error {
	var mangaData struct {
		Title       string   `json:"title"`
		Author      string   `json:"author"`
		Description string   `json:"description"`
		Price       float64  `json:"price"`
		Quantity    int      `json:"quantity"`
		Genres      []string `json:"genres"`
	}

	if err := c.BodyParser(&mangaData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	err := h.mangaService.CreateManga(mangaData.Title, mangaData.Author, mangaData.Description, mangaData.Price, mangaData.Quantity, mangaData.Genres)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Manga created successfully"})
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

	encUserId := c.Cookies("data")
	decUserId, err := helpers.Decrypt(encUserId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user credentials, try loggin in again"})
	}

	manga, err := h.mangaService.GetMangaByID(id, decUserId)
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

func (h MangaHandler) DeleteManga(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Manga ID is required",
		})
	}

	mongoId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Manga ID is invalid",
		})
	}

	err = h.mangaService.DeleteManga(mongoId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete manga",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Manga deleted successfully",
	})
}


func (h MangaHandler) PurchaseManga(c *fiber.Ctx) error {
	var request models.PurchaseRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	encUserId := c.Cookies("data")
	decUserId, err := helpers.Decrypt(encUserId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user credentials, try loggin in again"})
	}

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
	encUserId := c.Cookies("data")
	decUserId, err := helpers.Decrypt(encUserId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user credentials, try loggin in again"})
	}

	mangaId := c.Params("id")
	if mangaId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Manga ID is required"})
	}

	scoreStr := c.Query("score")
	if scoreStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Score is required"})
	}

	score, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Score must be a valid number"})
	}
	if score < 1 || score > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Score must be between 1 and 5"})
	}

	err = h.mangaService.RateManga(mangaId, decUserId, score)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to rate manga"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Manga rated successfully"})
}

func (h MangaHandler) RemoveMangaRating(c *fiber.Ctx) error {
	encUserId := c.Cookies("data")
	decUserId, err := helpers.Decrypt(encUserId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user credentials, try loggin in again"})
	}

	mangaId := c.Params("id")
	if mangaId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Manga ID is required"})
	}

	err = h.mangaService.RemoveMangaRating(mangaId, decUserId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove manga rating"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Manga rating removed successfully"})
}
