package main

import (
	"fmt"
	"manga_store/internal/databases"
	"manga_store/internal/helpers"
	"manga_store/internal/routers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	clientPort := helpers.GetEnv("CLIENT_PORT", "5173")

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:" + clientPort,
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept",
	}))

	databases.InitMongo()
	databases.InitNeo4j()
	databases.InitRedis()

	routers.NewAuthRouter().SetupRoutes(app)

	app.Use(AuthMiddleware())

	routers.NewMangaRouter().SetupRoutes(app)
	routers.NewUserRouter().SetupRoutes(app)

	port := helpers.GetEnv("PORT", "3000")
	app.Listen(fmt.Sprintf(":%s", port))
}

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		loggedIn := c.Cookies("loggedIn")
		data := c.Cookies("data")
		if loggedIn != "true" || data == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}
		return c.Next()
	}
}
