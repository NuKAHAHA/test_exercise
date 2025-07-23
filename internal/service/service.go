package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"awesomeProject1/internal/model"
)

type SubscriptionService struct {
	repo   repositorySubscription
	logger *slog.Logger
}

type repositorySubscription interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, sub *models.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, serviceName string) ([]models.Subscription, error)
	Aggregate(ctx context.Context, start time.Time, end time.Time, userID *uuid.UUID, serviceName *string) (int, error)
}

func NewSubscriptionService(repo repositorySubscription, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, serviceName string, price int, userID uuid.UUID, startDateStr string, endDateStr string) (*models.Subscription, error) {
	s.logger.InfoContext(ctx, "Creating subscription in service layer",
		slog.String("service_name", serviceName),
		slog.Int("price", price),
		slog.String("user_id", userID.String()),
		slog.String("start_date", startDateStr),
		slog.String("end_date", endDateStr))

	s.logger.DebugContext(ctx, "Parsing start date", slog.String("start_date", startDateStr))
	startDate, err := parseMonthYear(startDateStr)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to parse start date",
			slog.String("start_date", startDateStr),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}

	var endDate *time.Time
	if endDateStr != "" {
		s.logger.DebugContext(ctx, "Parsing end date", slog.String("end_date", endDateStr))
		ed, err := parseMonthYear(endDateStr)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to parse end date",
				slog.String("end_date", endDateStr),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}
		if ed.Before(startDate) {
			s.logger.ErrorContext(ctx, "End date is before start date",
				slog.Time("start_date", startDate),
				slog.Time("end_date", ed))
			return nil, errors.New("end_date must be after start_date")
		}
		endDate = &ed
	}

	subID := uuid.New()
	sub := &models.Subscription{
		ID:          subID,
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	s.logger.InfoContext(ctx, "Subscription model created, calling repository",
		slog.String("subscription_id", subID.String()))

	if err := s.repo.Create(ctx, sub); err != nil {
		s.logger.ErrorContext(ctx, "Repository failed to create subscription",
			slog.String("subscription_id", subID.String()),
			slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.InfoContext(ctx, "Successfully created subscription in service layer",
		slog.String("subscription_id", subID.String()))

	return sub, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	s.logger.InfoContext(ctx, "Retrieving subscription by ID in service layer",
		slog.String("subscription_id", id.String()))

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "Repository failed to get subscription",
			slog.String("subscription_id", id.String()),
			slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.InfoContext(ctx, "Successfully retrieved subscription in service layer",
		slog.String("subscription_id", id.String()),
		slog.String("service_name", sub.ServiceName))

	return sub, nil
}

func (s *SubscriptionService) Update(ctx context.Context, id uuid.UUID, serviceName string, price int, startDateStr string, endDateStr string) (*models.Subscription, error) {
	s.logger.InfoContext(ctx, "Updating subscription in service layer",
		slog.String("subscription_id", id.String()),
		slog.String("service_name", serviceName),
		slog.Int("price", price),
		slog.String("start_date", startDateStr),
		slog.String("end_date", endDateStr))

	s.logger.DebugContext(ctx, "Retrieving existing subscription for update",
		slog.String("subscription_id", id.String()))

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to retrieve subscription for update",
			slog.String("subscription_id", id.String()),
			slog.String("error", err.Error()))
		return nil, err
	}

	var updatedFields []string

	if serviceName != "" {
		sub.ServiceName = serviceName
		updatedFields = append(updatedFields, "service_name")
	}

	if price > 0 {
		sub.Price = price
		updatedFields = append(updatedFields, "price")
	}

	if startDateStr != "" {
		s.logger.DebugContext(ctx, "Parsing new start date", slog.String("start_date", startDateStr))
		startDate, err := parseMonthYear(startDateStr)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to parse new start date",
				slog.String("start_date", startDateStr),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("invalid start_date: %w", err)
		}
		sub.StartDate = startDate
		updatedFields = append(updatedFields, "start_date")
	}

	if endDateStr != "" {
		s.logger.DebugContext(ctx, "Parsing new end date", slog.String("end_date", endDateStr))
		endDate, err := parseMonthYear(endDateStr)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to parse new end date",
				slog.String("end_date", endDateStr),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}
		if endDate.Before(sub.StartDate) {
			s.logger.ErrorContext(ctx, "New end date is before start date",
				slog.Time("start_date", sub.StartDate),
				slog.Time("end_date", endDate))
			return nil, errors.New("end_date must be after start_date")
		}
		sub.EndDate = &endDate
		updatedFields = append(updatedFields, "end_date")
	} else {
		sub.EndDate = nil
		updatedFields = append(updatedFields, "end_date_cleared")
	}

	s.logger.InfoContext(ctx, "Fields to be updated",
		slog.String("subscription_id", id.String()),
		slog.Any("updated_fields", updatedFields))

	if err := s.repo.Update(ctx, sub); err != nil {
		s.logger.ErrorContext(ctx, "Repository failed to update subscription",
			slog.String("subscription_id", id.String()),
			slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.InfoContext(ctx, "Successfully updated subscription in service layer",
		slog.String("subscription_id", id.String()))

	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.InfoContext(ctx, "Deleting subscription in service layer",
		slog.String("subscription_id", id.String()))

	err := s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "Repository failed to delete subscription",
			slog.String("subscription_id", id.String()),
			slog.String("error", err.Error()))
		return err
	}

	s.logger.InfoContext(ctx, "Successfully deleted subscription in service layer",
		slog.String("subscription_id", id.String()))

	return nil
}

func (s *SubscriptionService) List(ctx context.Context, userID uuid.UUID, serviceName string) ([]models.Subscription, error) {
	s.logger.InfoContext(ctx, "Listing subscriptions in service layer",
		slog.String("user_id", userID.String()),
		slog.String("service_name", serviceName))

	subs, err := s.repo.List(ctx, userID, serviceName)
	if err != nil {
		s.logger.ErrorContext(ctx, "Repository failed to list subscriptions",
			slog.String("user_id", userID.String()),
			slog.String("service_name", serviceName),
			slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.InfoContext(ctx, "Successfully retrieved subscriptions list in service layer",
		slog.String("user_id", userID.String()),
		slog.String("service_name", serviceName),
		slog.Int("count", len(subs)))

	return subs, nil
}

func (s *SubscriptionService) Aggregate(ctx context.Context, startDateStr string, endDateStr string, userID *uuid.UUID, serviceName *string) (int, error) {
	var userIDStr string
	var serviceNameStr string

	if userID != nil {
		userIDStr = userID.String()
	}
	if serviceName != nil {
		serviceNameStr = *serviceName
	}

	s.logger.InfoContext(ctx, "Aggregating subscription costs in service layer",
		slog.String("start_date", startDateStr),
		slog.String("end_date", endDateStr),
		slog.String("user_id", userIDStr),
		slog.String("service_name", serviceNameStr))

	s.logger.DebugContext(ctx, "Parsing aggregation start date", slog.String("start_date", startDateStr))
	startDate, err := parseMonthYear(startDateStr)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to parse aggregation start date",
			slog.String("start_date", startDateStr),
			slog.String("error", err.Error()))
		return 0, fmt.Errorf("invalid start_date: %w", err)
	}

	s.logger.DebugContext(ctx, "Parsing aggregation end date", slog.String("end_date", endDateStr))
	endDate, err := parseMonthYear(endDateStr)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to parse aggregation end date",
			slog.String("end_date", endDateStr),
			slog.String("error", err.Error()))
		return 0, fmt.Errorf("invalid end_date: %w", err)
	}

	startPeriod := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	endPeriod := time.Date(endDate.Year(), endDate.Month()+1, 0, 23, 59, 59, 0, time.UTC)

	s.logger.DebugContext(ctx, "Calculated aggregation period",
		slog.Time("start_period", startPeriod),
		slog.Time("end_period", endPeriod))

	total, err := s.repo.Aggregate(ctx, startPeriod, endPeriod, userID, serviceName)
	if err != nil {
		s.logger.ErrorContext(ctx, "Repository failed to aggregate subscriptions",
			slog.Time("start_period", startPeriod),
			slog.Time("end_period", endPeriod),
			slog.String("user_id", userIDStr),
			slog.String("service_name", serviceNameStr),
			slog.String("error", err.Error()))
		return 0, err
	}

	s.logger.InfoContext(ctx, "Successfully completed aggregation in service layer",
		slog.Time("start_period", startPeriod),
		slog.Time("end_period", endPeriod),
		slog.String("user_id", userIDStr),
		slog.String("service_name", serviceNameStr),
		slog.Int("total", total))

	return total, nil
}

func parseMonthYear(dateStr string) (time.Time, error) {
	return time.Parse("01-2006", dateStr)
}
