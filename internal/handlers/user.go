package handlers

import (
	"manga_store/internal/helpers"
	"manga_store/internal/services"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler() UserHandler {
	return UserHandler{
		userService: services.NewUserService(),
	}
}

func (h UserHandler) GetRecsByPreferences(c *fiber.Ctx) error {
	encUserId := c.Cookies("data")
	userID, err := helpers.Decrypt(encUserId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user credentials, try loggin in again"})
	}
	recommendations, err := h.userService.GetRecsByPreferences(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get recommendations",
		})
	}
	return c.JSON(recommendations)
}

func (h UserHandler) GetRecsBySimilarUsers(c *fiber.Ctx) error {
	encUserId := c.Cookies("data")
	userID, err := helpers.Decrypt(encUserId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user credentials, try loggin in again"})
	}
	recommendations, err := h.userService.GetRecsBySimilarUsers(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get recommendations",
		})
	}
	return c.JSON(recommendations)
}

func (h UserHandler) DeleteUser(c *fiber.Ctx) error {
	encUserId := c.Cookies("data")
	decUserId, err := helpers.Decrypt(encUserId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user credentials, try loggin in again"})
	}

	userId, err := primitive.ObjectIDFromHex(decUserId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is invalid",
		})
	}

	err = h.userService.DeleteUser(userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

func (h UserHandler) RestoreUser(c *fiber.Ctx) error {
	isAdmin := c.Cookies("isAdmin")
	if isAdmin != "true" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	id := c.Params("id")
	if isAdmin == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	userId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is invalid",
		})
	}

	err = h.userService.RestoreUser(userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User restored successfully",
	})
}
