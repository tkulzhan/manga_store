package main

import (
	"fmt"
	"manga_store/internal/databases"
	"manga_store/internal/helpers"
	"manga_store/internal/routers"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	databases.InitMongo()
	databases.InitNeo4j()
	databases.InitRedis()

	routers.NewAuthRouter().SetupRoutes(app)
	routers.NewMangaRouter().SetupRoutes(app)
	routers.NewRecsRouter().SetupRoutes(app)

	port := helpers.GetEnv("PORT", "3000")
	app.Listen(fmt.Sprintf(":%s", port))
}