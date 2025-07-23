package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"awesomeProject1/internal/model"
)

type SubscriptionRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewSubscriptionRepository(db *gorm.DB, logger *slog.Logger) *SubscriptionRepository {
	return &SubscriptionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	r.logger.InfoContext(ctx, "Creating new subscription in repository",
		slog.String("subscription_id", sub.ID.String()),
		slog.String("service_name", sub.ServiceName),
		slog.String("user_id", sub.UserID.String()))

	start := time.Now()
	err := r.db.WithContext(ctx).Create(sub).Error

	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to create subscription in database",
			slog.String("subscription_id", sub.ID.String()),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))
		return err
	}

	r.logger.InfoContext(ctx, "Successfully created subscription in database",
		slog.String("subscription_id", sub.ID.String()),
		slog.Duration("duration", time.Since(start)))

	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	r.logger.InfoContext(ctx, "Retrieving subscription by ID from repository",
		slog.String("subscription_id", id.String()))

	start := time.Now()
	var sub models.Subscription
	err := r.db.WithContext(ctx).First(&sub, "id = ?", id).Error

	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to retrieve subscription from database",
			slog.String("subscription_id", id.String()),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))
		return nil, err
	}

	r.logger.InfoContext(ctx, "Successfully retrieved subscription from database",
		slog.String("subscription_id", id.String()),
		slog.String("service_name", sub.ServiceName),
		slog.Duration("duration", time.Since(start)))

	return &sub, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *models.Subscription) error {
	r.logger.InfoContext(ctx, "Updating subscription in repository",
		slog.String("subscription_id", sub.ID.String()),
		slog.String("service_name", sub.ServiceName))

	start := time.Now()
	err := r.db.WithContext(ctx).Save(sub).Error

	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to update subscription in database",
			slog.String("subscription_id", sub.ID.String()),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))
		return err
	}

	r.logger.InfoContext(ctx, "Successfully updated subscription in database",
		slog.String("subscription_id", sub.ID.String()),
		slog.Duration("duration", time.Since(start)))

	return nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.InfoContext(ctx, "Deleting subscription from repository",
		slog.String("subscription_id", id.String()))

	start := time.Now()
	err := r.db.WithContext(ctx).Delete(&models.Subscription{}, "id = ?", id).Error

	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to delete subscription from database",
			slog.String("subscription_id", id.String()),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))
		return err
	}

	r.logger.InfoContext(ctx, "Successfully deleted subscription from database",
		slog.String("subscription_id", id.String()),
		slog.Duration("duration", time.Since(start)))

	return nil
}

func (r *SubscriptionRepository) List(ctx context.Context, userID uuid.UUID, serviceName string) ([]models.Subscription, error) {
	r.logger.InfoContext(ctx, "Listing subscriptions from repository",
		slog.String("user_id", userID.String()),
		slog.String("service_name", serviceName))

	start := time.Now()
	var subs []models.Subscription
	query := r.db.WithContext(ctx)

	if userID != uuid.Nil {
		query = query.Where("user_id = ?", userID)
		r.logger.DebugContext(ctx, "Applied user_id filter", slog.String("user_id", userID.String()))
	}

	if serviceName != "" {
		query = query.Where("service_name = ?", serviceName)
		r.logger.DebugContext(ctx, "Applied service_name filter", slog.String("service_name", serviceName))
	}

	err := query.Find(&subs).Error

	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to list subscriptions from database",
			slog.String("user_id", userID.String()),
			slog.String("service_name", serviceName),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)))
		return nil, err
	}

	r.logger.InfoContext(ctx, "Successfully retrieved subscriptions list from database",
		slog.String("user_id", userID.String()),
		slog.String("service_name", serviceName),
		slog.Int("count", len(subs)),
		slog.Duration("duration", time.Since(start)))

	return subs, nil
}

func (r *SubscriptionRepository) Aggregate(ctx context.Context, start time.Time, end time.Time, userID *uuid.UUID, serviceName *string) (int, error) {
	var userIDStr string
	var serviceNameStr string

	if userID != nil {
		userIDStr = userID.String()
	}
	if serviceName != nil {
		serviceNameStr = *serviceName
	}

	r.logger.InfoContext(ctx, "Aggregating subscription costs from repository",
		slog.Time("start_date", start),
		slog.Time("end_date", end),
		slog.String("user_id", userIDStr),
		slog.String("service_name", serviceNameStr))

	queryStart := time.Now()
	db := r.db.WithContext(ctx).Model(&models.Subscription{}).
		Select("COALESCE(SUM(price), 0)").
		Where("start_date <= ?", end).
		Where("(end_date >= ? OR end_date IS NULL)", start)

	if userID != nil {
		db = db.Where("user_id = ?", *userID)
		r.logger.DebugContext(ctx, "Applied user_id filter for aggregation", slog.String("user_id", userID.String()))
	}

	if serviceName != nil {
		db = db.Where("service_name = ?", *serviceName)
		r.logger.DebugContext(ctx, "Applied service_name filter for aggregation", slog.String("service_name", *serviceName))
	}

	var total int
	row := db.Row()
	if err := row.Scan(&total); err != nil {
		r.logger.ErrorContext(ctx, "Aggregation query failed",
			slog.Time("start_date", start),
			slog.Time("end_date", end),
			slog.String("user_id", userIDStr),
			slog.String("service_name", serviceNameStr),
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(queryStart)))
		return 0, fmt.Errorf("aggregation query failed: %w", err)
	}

	r.logger.InfoContext(ctx, "Successfully completed aggregation query",
		slog.Time("start_date", start),
		slog.Time("end_date", end),
		slog.String("user_id", userIDStr),
		slog.String("service_name", serviceNameStr),
		slog.Int("total", total),
		slog.Duration("duration", time.Since(queryStart)))

	return total, nil
}
