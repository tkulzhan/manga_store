package handlers

import (
	"manga_store/internal/helpers"
	"manga_store/internal/services"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler() AuthHandler {
	return AuthHandler{
		authService: services.NewAuthService(),
	}
}

func (h AuthHandler) Register(c *fiber.Ctx) error {
	var registerData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&registerData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	err := h.authService.Register(registerData.Email, registerData.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User registered successfully"})
}

func (h AuthHandler) Login(c *fiber.Ctx) error {
	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&loginData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	user, err := h.authService.Login(loginData.Email, loginData.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "loggedIn",
		Value:    "true",
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Strict",
	})

	hashedID := helpers.Encrypt(user.ID)

	c.Cookie(&fiber.Cookie{
		Name:     "data",
		Value:    hashedID,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Strict",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Logged in successfully", "user": user})
}

func (h AuthHandler) Logout(c *fiber.Ctx) error {
	err := h.authService.Logout()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to log out"})
	}

	c.ClearCookie()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Logged out successfully"})
}
