package repository

import (
	"context"
	"errors"

	"github.com/cristiano-pacheco/go-otel/trace"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/model"
	"github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/database"
	"gorm.io/gorm"
)

type NotificationRepositoryI interface {
	FindByID(ctx context.Context, notificationID uint64) (model.NotificationModel, error)
	FindByMonitorID(ctx context.Context, monitorID uint64) ([]model.NotificationModel, error)
	Create(ctx context.Context, notification model.NotificationModel) (model.NotificationModel, error)
	Update(ctx context.Context, notification model.NotificationModel) (model.NotificationModel, error)
}

type NotificationRepository struct {
	*database.PingoDB
}

var _ NotificationRepositoryI = (*NotificationRepository)(nil)

func NewNotificationRepository(db *database.PingoDB) *NotificationRepository {
	return &NotificationRepository{db}
}

func (r *NotificationRepository) FindByID(ctx context.Context, notificationID uint64) (model.NotificationModel, error) {
	ctx, otelSpan := trace.Span(ctx, "NotificationRepository.FindByID")
	defer otelSpan.End()

	notification, err := gorm.G[model.NotificationModel](r.DB).
		Where("id = ?", notificationID).
		Limit(1).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.NotificationModel{}, errs.ErrRecordNotFound
		}
		return model.NotificationModel{}, err
	}
	return notification, nil
}

func (r *NotificationRepository) FindByMonitorID(
	ctx context.Context,
	monitorID uint64,
) ([]model.NotificationModel, error) {
	ctx, otelSpan := trace.Span(ctx, "NotificationRepository.FindByMonitorID")
	defer otelSpan.End()

	notifications, err := gorm.G[model.NotificationModel](r.DB).
		Where("http_monitor_id = ?", monitorID).
		Order("sent_at DESC").
		Find(ctx)

	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *NotificationRepository) Create(
	ctx context.Context,
	notification model.NotificationModel,
) (model.NotificationModel, error) {
	ctx, otelSpan := trace.Span(ctx, "NotificationRepository.Create")
	defer otelSpan.End()

	err := gorm.G[model.NotificationModel](r.DB).Create(ctx, &notification)
	return notification, err
}

func (r *NotificationRepository) Update(
	ctx context.Context,
	notification model.NotificationModel,
) (model.NotificationModel, error) {
	ctx, otelSpan := trace.Span(ctx, "NotificationRepository.Update")
	defer otelSpan.End()

	_, err := gorm.G[model.NotificationModel](r.DB).Updates(ctx, notification)
	if err != nil {
		return model.NotificationModel{}, err
	}
	return notification, nil
}
