package subscriptions

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"test-em/internal/logger"
)

type SubscriptionHandler struct {
	db *gorm.DB
}

func NewSubscriptionHandler(db *gorm.DB) *SubscriptionHandler {
	return &SubscriptionHandler{db: db}
}

func RegisterNewRoutes(app *fiber.App, db *gorm.DB) {
	logger.Info("Registering subscription routes")

	h := NewSubscriptionHandler(db)

	app.Route("/api", func(api fiber.Router) {
		api.Route("/subscriptions", func(router fiber.Router) {
			logger.Debug("Registering subscription endpoints")
			router.Get("/", h.ListSubscriptions)
			router.Post("/", h.CreateSubscription)
			router.Get("/:id", h.GetSubscription)
			router.Put("/:id", h.UpdateSubscription)
			router.Delete("/:id", h.DeleteSubscription)
			router.Get("/sum/:userid", h.CountSubscriptionSum)
		})
	})

	logger.Info("Subscription routes registered successfully")
}
