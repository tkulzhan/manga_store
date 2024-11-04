package handlers

import (
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

func (h UserHandler) GetUserByPreferences(c *fiber.Ctx) error {
	return h.userService.GetUserByPreferences()
}

func (h UserHandler) GetUserBySimilarUsers(c *fiber.Ctx) error {
	return h.userService.GetUserBySimilarUsers()
}

func (h UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
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

