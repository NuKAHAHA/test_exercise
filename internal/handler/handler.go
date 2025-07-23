package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"awesomeProject1/internal/model"
)

type SubscriptionHandler struct {
	service SubscriptionService
	logger  *slog.Logger
}

type SubscriptionService interface {
	Create(ctx context.Context, serviceName string, price int, userID uuid.UUID, startDateStr string, endDateStr string) (*models.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, serviceName string, price int, startDateStr string, endDateStr string) (*models.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, serviceName string) ([]models.Subscription, error)
	Aggregate(ctx context.Context, startDateStr string, endDateStr string, userID *uuid.UUID, serviceName *string) (int, error)
}

func NewSubscriptionHandler(service SubscriptionService, logger *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SubscriptionHandler) Create(c *gin.Context) {
	start := time.Now()
	requestID := uuid.New().String()

	h.logger.Info("Starting subscription creation",
		slog.String("request_id", requestID),
		slog.String("method", "Create"),
		slog.String("client_ip", c.ClientIP()),
		slog.String("user_agent", c.GetHeader("User-Agent")))

	var req struct {
		ServiceName string    `json:"service_name" binding:"required"`
		Price       int       `json:"price" binding:"required,gt=0"`
		UserID      uuid.UUID `json:"user_id" binding:"required"`
		StartDate   string    `json:"start_date" binding:"required"`
		EndDate     string    `json:"end_date,omitempty"`
	}

	h.logger.Debug("Attempting to bind JSON request",
		slog.String("request_id", requestID))

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON request",
			slog.String("request_id", requestID),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Successfully bound request data",
		slog.String("request_id", requestID),
		slog.String("service_name", req.ServiceName),
		slog.Int("price", req.Price),
		slog.String("user_id", req.UserID.String()),
		slog.String("start_date", req.StartDate),
		slog.String("end_date", req.EndDate))

	h.logger.Debug("Calling service.Create",
		slog.String("request_id", requestID))

	sub, err := h.service.Create(c.Request.Context(), req.ServiceName, req.Price, req.UserID, req.StartDate, req.EndDate)
	if err != nil {
		h.logger.Error("Service.Create failed",
			slog.String("request_id", requestID),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Successfully created subscription",
		slog.String("request_id", requestID),
		slog.String("subscription_id", sub.ID.String()),
		slog.Duration("duration", time.Since(start)))

	c.JSON(http.StatusCreated, sub)
}

func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	start := time.Now()
	requestID := uuid.New().String()
	idParam := c.Param("id")

	h.logger.Info("Starting subscription retrieval",
		slog.String("request_id", requestID),
		slog.String("method", "GetByID"),
		slog.String("id_param", idParam),
		slog.String("client_ip", c.ClientIP()))

	h.logger.Debug("Attempting to parse UUID",
		slog.String("request_id", requestID),
		slog.String("id_param", idParam))

	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Error("Failed to parse UUID",
			slog.String("request_id", requestID),
			slog.String("id_param", idParam),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	h.logger.Debug("Successfully parsed UUID, calling service.GetByID",
		slog.String("request_id", requestID),
		slog.String("subscription_id", id.String()))

	sub, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Warn("Subscription not found",
				slog.String("request_id", requestID),
				slog.String("subscription_id", id.String()),
				slog.Duration("duration", time.Since(start)))

			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}

		h.logger.Error("Service.GetByID failed",
			slog.String("request_id", requestID),
			slog.String("subscription_id", id.String()),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription"})
		return
	}

	h.logger.Info("Successfully retrieved subscription",
		slog.String("request_id", requestID),
		slog.String("subscription_id", id.String()),
		slog.String("service_name", sub.ServiceName),
		slog.Duration("duration", time.Since(start)))

	c.JSON(http.StatusOK, sub)
}

func (h *SubscriptionHandler) Update(c *gin.Context) {
	start := time.Now()
	requestID := uuid.New().String()
	idParam := c.Param("id")

	h.logger.Info("Starting subscription update",
		slog.String("request_id", requestID),
		slog.String("method", "Update"),
		slog.String("id_param", idParam),
		slog.String("client_ip", c.ClientIP()))

	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Error("Failed to parse UUID",
			slog.String("request_id", requestID),
			slog.String("id_param", idParam),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	var req struct {
		ServiceName string `json:"service_name,omitempty"`
		Price       int    `json:"price,omitempty"`
		StartDate   string `json:"start_date,omitempty"`
		EndDate     string `json:"end_date,omitempty"`
	}

	h.logger.Debug("Attempting to bind JSON request for update",
		slog.String("request_id", requestID),
		slog.String("subscription_id", id.String()))

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON request for update",
			slog.String("request_id", requestID),
			slog.String("subscription_id", id.String()),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Successfully bound update request data",
		slog.String("request_id", requestID),
		slog.String("subscription_id", id.String()),
		slog.String("service_name", req.ServiceName),
		slog.Int("price", req.Price),
		slog.String("start_date", req.StartDate),
		slog.String("end_date", req.EndDate))

	h.logger.Debug("Calling service.Update",
		slog.String("request_id", requestID),
		slog.String("subscription_id", id.String()))

	sub, err := h.service.Update(c.Request.Context(), id, req.ServiceName, req.Price, req.StartDate, req.EndDate)
	if err != nil {
		h.logger.Error("Service.Update failed",
			slog.String("request_id", requestID),
			slog.String("subscription_id", id.String()),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Successfully updated subscription",
		slog.String("request_id", requestID),
		slog.String("subscription_id", id.String()),
		slog.Duration("duration", time.Since(start)))

	c.JSON(http.StatusOK, sub)
}

func (h *SubscriptionHandler) Delete(c *gin.Context) {
	start := time.Now()
	requestID := uuid.New().String()
	idParam := c.Param("id")

	h.logger.Info("Starting subscription deletion",
		slog.String("request_id", requestID),
		slog.String("method", "Delete"),
		slog.String("id_param", idParam),
		slog.String("client_ip", c.ClientIP()))

	id, err := uuid.Parse(idParam)
	if err != nil {
		h.logger.Error("Failed to parse UUID for deletion",
			slog.String("request_id", requestID),
			slog.String("id_param", idParam),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	h.logger.Debug("Calling service.Delete",
		slog.String("request_id", requestID),
		slog.String("subscription_id", id.String()))

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Warn("Attempted to delete non-existent subscription",
				slog.String("request_id", requestID),
				slog.String("subscription_id", id.String()),
				slog.Duration("duration", time.Since(start)))

			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}

		h.logger.Error("Service.Delete failed",
			slog.String("request_id", requestID),
			slog.String("subscription_id", id.String()),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete subscription"})
		return
	}

	h.logger.Info("Successfully deleted subscription",
		slog.String("request_id", requestID),
		slog.String("subscription_id", id.String()),
		slog.Duration("duration", time.Since(start)))

	c.Status(http.StatusNoContent)
}

func (h *SubscriptionHandler) List(c *gin.Context) {
	start := time.Now()
	requestID := uuid.New().String()
	userIDParam := c.Query("user_id")
	serviceName := c.Query("service_name")

	h.logger.Info("Starting subscription listing",
		slog.String("request_id", requestID),
		slog.String("method", "List"),
		slog.String("user_id_param", userIDParam),
		slog.String("service_name", serviceName),
		slog.String("client_ip", c.ClientIP()))

	userID, parseErr := uuid.Parse(userIDParam)
	if parseErr != nil && userIDParam != "" {
		h.logger.Warn("Invalid user_id parameter provided",
			slog.String("request_id", requestID),
			slog.String("user_id_param", userIDParam),
			slog.String("parse_error", parseErr.Error()))
	}

	h.logger.Debug("Calling service.List",
		slog.String("request_id", requestID),
		slog.String("user_id", userID.String()),
		slog.String("service_name", serviceName))

	subs, err := h.service.List(c.Request.Context(), userID, serviceName)
	if err != nil {
		h.logger.Error("Service.List failed",
			slog.String("request_id", requestID),
			slog.String("user_id", userID.String()),
			slog.String("service_name", serviceName),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list subscriptions"})
		return
	}

	h.logger.Info("Successfully retrieved subscriptions list",
		slog.String("request_id", requestID),
		slog.Int("count", len(subs)),
		slog.String("user_id", userID.String()),
		slog.String("service_name", serviceName),
		slog.Duration("duration", time.Since(start)))

	c.JSON(http.StatusOK, subs)
}

func (h *SubscriptionHandler) Aggregate(c *gin.Context) {
	start := time.Now()
	requestID := uuid.New().String()

	h.logger.Info("Starting subscription aggregation",
		slog.String("request_id", requestID),
		slog.String("method", "Aggregate"),
		slog.String("client_ip", c.ClientIP()))

	var req struct {
		StartDate   string     `json:"start_date" binding:"required"`
		EndDate     string     `json:"end_date" binding:"required"`
		UserID      *uuid.UUID `json:"user_id,omitempty"`
		ServiceName *string    `json:"service_name,omitempty"`
	}

	h.logger.Debug("Attempting to bind JSON request for aggregation",
		slog.String("request_id", requestID))

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON request for aggregation",
			slog.String("request_id", requestID),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userIDStr string
	var serviceNameStr string

	if req.UserID != nil {
		userIDStr = req.UserID.String()
	}
	if req.ServiceName != nil {
		serviceNameStr = *req.ServiceName
	}

	h.logger.Info("Successfully bound aggregation request data",
		slog.String("request_id", requestID),
		slog.String("start_date", req.StartDate),
		slog.String("end_date", req.EndDate),
		slog.String("user_id", userIDStr),
		slog.String("service_name", serviceNameStr))

	h.logger.Debug("Calling service.Aggregate",
		slog.String("request_id", requestID))

	total, err := h.service.Aggregate(c.Request.Context(),
		req.StartDate,
		req.EndDate,
		req.UserID,
		req.ServiceName,
	)

	if err != nil {
		h.logger.Error("Service.Aggregate failed",
			slog.String("request_id", requestID),
			slog.String("start_date", req.StartDate),
			slog.String("end_date", req.EndDate),
			slog.String("user_id", userIDStr),
			slog.String("service_name", serviceNameStr),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Successfully calculated aggregation",
		slog.String("request_id", requestID),
		slog.Int("total", total),
		slog.String("start_date", req.StartDate),
		slog.String("end_date", req.EndDate),
		slog.String("user_id", userIDStr),
		slog.String("service_name", serviceNameStr),
		slog.Duration("duration", time.Since(start)))

	c.JSON(http.StatusOK, gin.H{"total": total})
}
