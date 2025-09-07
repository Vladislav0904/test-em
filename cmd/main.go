package main

import (
	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"test-em/internal/config"
	"test-em/internal/database"
	"test-em/internal/logger"
	"test-em/internal/subscriptions"
)

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg)
	logger.Info("Starting subscriptions API service")

	logger.Info("Connecting to database", "host", cfg.DbHost, "port", cfg.DbPort, "database", cfg.DbName)
	db := database.LoadDB(cfg)
	logger.Info("Database connection established successfully")

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.WithError(err).Error("Request processing error", "path", c.Path(), "method", c.Method())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
		},
	})

	app.Use(recover.New())
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	logger.Info("Registering API routes")
	subscriptions.RegisterNewRoutes(app, db)
	logger.Info("Starting HTTP server on port 8080")
	if err := app.Listen(":8080"); err != nil {
		logger.WithError(err).Fatal("Failed to start the server")
	}
}
