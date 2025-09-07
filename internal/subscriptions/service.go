package subscriptions

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"test-em/internal/logger"
	"time"
)

type createSubscriptionRequest struct {
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type subscriptionResponse struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

func (h *SubscriptionHandler) CreateSubscription(ctx *fiber.Ctx) error {
	logger.Info("Creating new subscription", "method", "POST", "path", "/api/subscriptions/")

	var req createSubscriptionRequest
	if err := ctx.BodyParser(&req); err != nil {
		logger.WithError(err).Error("Failed to parse request body")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	logger.Debug("Subscription creation request", "service_name", req.ServiceName, "price", req.Price, "user_id", req.UserID, "start_date", req.StartDate, "end_date", req.EndDate)

	req.ServiceName = strings.TrimSpace(req.ServiceName)
	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		logger.WithError(err).Error("Failed to parse start date", "start_date", req.StartDate)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	var endDatePtr *time.Time
	if strings.TrimSpace(req.EndDate) != "" {
		endDate, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			logger.WithError(err).Error("Failed to parse end date", "end_date", req.EndDate)
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}
		endDatePtr = &endDate
	}

	subscription := Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      uuid.Must(uuid.Parse(req.UserID)),
		StartDate:   startDate,
		EndDate:     endDatePtr,
	}

	logger.Debug("Creating subscription in database", "subscription_id", subscription.ID, "user_id", subscription.UserID)
	if err := h.db.Create(&subscription).Error; err != nil {
		logger.WithError(err).Error("Failed to create subscription in database", "subscription_id", subscription.ID)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
	}

	logger.Info("Subscription created successfully", "subscription_id", subscription.ID, "user_id", subscription.UserID, "service_name", subscription.ServiceName)
	return ctx.Status(fiber.StatusCreated).JSON(subscriptionResponse{
		ID:          subscription.ID,
		ServiceName: subscription.ServiceName,
		Price:       subscription.Price,
		UserID:      subscription.UserID,
		StartDate:   subscription.StartDate,
		EndDate:     subscription.EndDate,
	})
}

// DeleteSubscription removes a subscription by id
func (h *SubscriptionHandler) DeleteSubscription(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	logger.Info("Deleting subscription", "method", "DELETE", "path", "/api/subscriptions/"+idParam, "subscription_id", idParam)

	id, err := uuid.Parse(idParam)
	if err != nil {
		logger.WithError(err).Error("Invalid subscription ID format", "subscription_id", idParam)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	logger.Debug("Deleting subscription from database", "subscription_id", id)
	if err := h.db.Delete(&Subscription{}, "id = ?", id).Error; err != nil {
		logger.WithError(err).Error("Failed to delete subscription from database", "subscription_id", id)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
	}

	logger.Info("Subscription deleted successfully", "subscription_id", id)
	return ctx.SendStatus(fiber.StatusNoContent)
}

// ListSubscriptions returns paginated list, optionally filtered by user_id and service_name
func (h *SubscriptionHandler) ListSubscriptions(ctx *fiber.Ctx) error {
	logger.Info("Listing subscriptions", "method", "GET", "path", "/api/subscriptions/")

	page := 1
	limit := 10

	if pStr := ctx.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}
	if lStr := ctx.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit
	userID := ctx.Query("user_id")
	serviceName := ctx.Query("service_name")

	logger.Debug("List subscriptions request", "page", page, "limit", limit, "user_id", userID, "service_name", serviceName)

	var items []Subscription
	var total int64

	baseQuery := h.db.Model(&Subscription{})

	if v := strings.TrimSpace(userID); v != "" {
		if _, err := uuid.Parse(v); err != nil {
			logger.WithError(err).Error("Invalid user_id format", "user_id", v)
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
		}
		baseQuery = baseQuery.Where("user_id = ?", v)
	}
	if v := strings.TrimSpace(serviceName); v != "" {
		baseQuery = baseQuery.Where("service_name = ?", v)
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		logger.WithError(err).Error("Failed to count subscriptions")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
	}

	if err := baseQuery.Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		logger.WithError(err).Error("Failed to fetch subscriptions")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
	}

	logger.Info("Subscriptions listed successfully", "page", page, "limit", limit, "offset", offset, "total", total, "items_count", len(items))

	totalPages := (total + int64(limit) - 1) / int64(limit)
	hasNext := int64(page) < totalPages
	hasPrev := page > 1

	return ctx.JSON(fiber.Map{
		"data": items,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
}

// UpdateSubscription updates fields of a subscription
func (h *SubscriptionHandler) UpdateSubscription(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	logger.Info("Updating subscription", "method", "PUT", "path", "/api/subscriptions/"+idParam, "subscription_id", idParam)

	id, err := uuid.Parse(idParam)
	if err != nil {
		logger.WithError(err).Error("Invalid subscription ID format", "subscription_id", idParam)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req createSubscriptionRequest
	if err := ctx.BodyParser(&req); err != nil {
		logger.WithError(err).Error("Failed to parse request body")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	logger.Debug("Subscription update request", "subscription_id", id, "updates", req)

	updates := map[string]interface{}{}
	if s := strings.TrimSpace(req.ServiceName); s != "" {
		updates["service_name"] = s
	}
	if req.Price != 0 {
		updates["price"] = req.Price
	}
	if u := strings.TrimSpace(req.UserID); u != "" {
		if _, err := uuid.Parse(u); err != nil {
			logger.WithError(err).Error("Invalid user_id format", "user_id", u)
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
		}
		updates["user_id"] = u
	}
	if s := strings.TrimSpace(req.StartDate); s != "" {
		t, err := time.Parse("01-2006", s)
		if err != nil {
			logger.WithError(err).Error("Invalid start_date format", "start_date", s)
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid start_date"})
		}
		updates["start_date"] = t
	}
	if s := strings.TrimSpace(req.EndDate); s != "" {
		t, err := time.Parse("01-2006", s)
		if err != nil {
			logger.WithError(err).Error("Invalid end_date format", "end_date", s)
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid end_date"})
		}
		updates["end_date"] = t
	}

	if len(updates) == 0 {
		logger.Warn("No updates provided", "subscription_id", id)
		return ctx.SendStatus(fiber.StatusNoContent)
	}

	logger.Debug("Updating subscription in database", "subscription_id", id, "updates", updates)
	if err := h.db.Model(&Subscription{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		logger.WithError(err).Error("Failed to update subscription in database", "subscription_id", id)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
	}

	var item Subscription
	if err := h.db.First(&item, "id = ?", id).Error; err != nil {
		logger.WithError(err).Error("Failed to fetch updated subscription", "subscription_id", id)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
	}

	logger.Info("Subscription updated successfully", "subscription_id", id)
	return ctx.JSON(item)
}

// GetSubscription returns single subscription by id
func (h *SubscriptionHandler) GetSubscription(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	logger.Info("Getting subscription", "method", "GET", "path", "/api/subscriptions/"+idParam, "subscription_id", idParam)

	id, err := uuid.Parse(idParam)
	if err != nil {
		logger.WithError(err).Error("Invalid subscription ID format", "subscription_id", idParam)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	logger.Debug("Fetching subscription from database", "subscription_id", id)
	var item Subscription
	if err := h.db.First(&item, "id = ?", id).Error; err != nil {
		logger.WithError(err).Warn("Subscription not found", "subscription_id", id)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}

	logger.Info("Subscription retrieved successfully", "subscription_id", id, "user_id", item.UserID, "service_name", item.ServiceName)
	return ctx.JSON(item)
}

// CountSubscriptionSum returns sum of prices within start-end period for a user, optional service filter
func (h *SubscriptionHandler) CountSubscriptionSum(ctx *fiber.Ctx) error {
	userid := strings.TrimSpace(ctx.Params("userid"))
	logger.Info("Calculating subscription sum", "method", "GET", "path", "/api/subscriptions/sum/"+userid, "user_id", userid)

	if userid == "" {
		logger.Error("Missing userid parameter")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "userid is required"})
	}
	if _, err := uuid.Parse(userid); err != nil {
		logger.WithError(err).Error("Invalid userid format", "userid", userid)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid userid"})
	}

	start := strings.TrimSpace(ctx.Query("start"))
	end := strings.TrimSpace(ctx.Query("end"))
	serviceName := strings.TrimSpace(ctx.Query("service_name"))

	logger.Debug("Sum calculation request", "user_id", userid, "start", start, "end", end, "service_name", serviceName)

	if start == "" || end == "" {
		logger.Error("Missing start or end parameters", "start", start, "end", end)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "start and end are required"})
	}

	periodStart, err := time.Parse("01-2006", start)
	if err != nil {
		logger.WithError(err).Error("Invalid start date format", "start", start)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid start"})
	}
	periodEnd, err := time.Parse("01-2006", end)
	if err != nil {
		logger.WithError(err).Error("Invalid end date format", "end", end)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid end"})
	}

	logger.Debug("Fetching overlapping subscriptions", "user_id", userid, "period_start", periodStart, "period_end", periodEnd)
	var items []Subscription
	q := h.db.Where("user_id = ? AND start_date <= ? AND (end_date IS NULL OR end_date >= ?)", userid, periodEnd, periodStart)
	if serviceName != "" {
		q = q.Where("service_name = ?", serviceName)
	}
	if err := q.Find(&items).Error; err != nil {
		logger.WithError(err).Error("Failed to fetch subscriptions for sum calculation")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
	}

	logger.Debug("Calculating sum for overlapping subscriptions", "subscriptions_count", len(items))
	sum := 0
	for _, s := range items {
		start := maxTime(s.StartDate, periodStart)
		endTime := periodEnd
		if s.EndDate != nil && s.EndDate.Before(endTime) {
			endTime = *s.EndDate
		}
		if start.After(endTime) {
			continue
		}
		months := diffMonthsInclusive(start, endTime)
		contribution := s.Price * months
		sum += contribution
		logger.Debug("Subscription contribution", "service_name", s.ServiceName, "price", s.Price, "months", months, "contribution", contribution)
	}

	logger.Info("Sum calculated successfully", "user_id", userid, "total_sum", sum, "subscriptions_processed", len(items))
	return ctx.JSON(fiber.Map{"sum": sum})
}

func diffMonthsInclusive(a, b time.Time) int {
	y1, m1, _ := a.Date()
	y2, m2, _ := b.Date()
	return (y2-y1)*12 + int(m2-m1) + 1
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
