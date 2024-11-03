package main

import (
	"manga_store/internal/routers"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	routers.NewAuthRouter().SetupRoutes(app)
	routers.NewMangaRouter().SetupRoutes(app)
	routers.NewRecsRouter().SetupRoutes(app)

	port := ":3000"
	app.Listen(port)
}