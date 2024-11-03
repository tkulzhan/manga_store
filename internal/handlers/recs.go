package handlers

import (
	"manga_store/internal/services"

	"github.com/gofiber/fiber/v2"
)

type RecsHandler struct {
	RecsService services.RecsService
}

func NewRecsHandler() RecsHandler {
	return RecsHandler{
		RecsService: services.NewRecsService(),
	}
}

func (h RecsHandler) GetRecsByPreferences(c *fiber.Ctx) error {
	return h.RecsService.GetRecsByPreferences()
}

func (h RecsHandler) GetRecsBySimilarUsers(c *fiber.Ctx) error {
	return h.RecsService.GetRecsBySimilarUsers()
}
