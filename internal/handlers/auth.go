package handlers

import (
	"manga_store/internal/services"

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
	return h.authService.Register()
}

func (h AuthHandler) Login(c *fiber.Ctx) error {
	return h.authService.Login()
}

func (h AuthHandler) Logout(c *fiber.Ctx) error {
	return h.authService.Logout()
}
