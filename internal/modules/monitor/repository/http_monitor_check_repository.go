package repository

import (
	"context"
	"errors"
	"time"

	"github.com/cristiano-pacheco/go-otel/trace"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/model"
	"github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/database"
	"gorm.io/gorm"
)

type HTTPMonitorCheckRepository interface {
	FindByID(ctx context.Context, checkID uint64) (model.HTTPMonitorCheckModel, error)
	FindAll(ctx context.Context, monitorID uint64, from, to *time.Time, page, pageSize int) ([]model.HTTPMonitorCheckModel, int64, error)
	Create(ctx context.Context, check model.HTTPMonitorCheckModel) (model.HTTPMonitorCheckModel, error)
}

type httpMonitorCheckRepository struct {
	*database.PingoDB
}

func NewHTTPMonitorCheckRepository(db *database.PingoDB) HTTPMonitorCheckRepository {
	return &httpMonitorCheckRepository{db}
}

func (r *httpMonitorCheckRepository) FindByID(ctx context.Context, checkID uint64) (model.HTTPMonitorCheckModel, error) {
	ctx, otelSpan := trace.StartSpan(ctx, "HTTPMonitorCheckRepository.FindByID")
	defer otelSpan.End()

	check, err := gorm.G[model.HTTPMonitorCheckModel](r.DB).
		Where("id = ?", checkID).
		Limit(1).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.HTTPMonitorCheckModel{}, errs.ErrRecordNotFound
		}
		return model.HTTPMonitorCheckModel{}, err
	}
	return check, nil
}

func (r *httpMonitorCheckRepository) FindAll(ctx context.Context, monitorID uint64, from, to *time.Time, page, pageSize int) ([]model.HTTPMonitorCheckModel, int64, error) {
	ctx, otelSpan := trace.StartSpan(ctx, "HTTPMonitorCheckRepository.FindAll")
	defer otelSpan.End()

	// Calculate offset
	offset := (page - 1) * pageSize

	// Build base query
	query := r.DB.Model(&model.HTTPMonitorCheckModel{}).
		Where("http_monitor_id = ?", monitorID)

	// Add optional date range filters
	if from != nil {
		query = query.Where("checked_at >= ?", *from)
	}
	if to != nil {
		query = query.Where("checked_at <= ?", *to)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Build query for paginated results
	checksQuery := gorm.G[model.HTTPMonitorCheckModel](r.DB).
		Where("http_monitor_id = ?", monitorID)

	// Add optional date range filters
	if from != nil {
		checksQuery = checksQuery.Where("checked_at >= ?", *from)
	}
	if to != nil {
		checksQuery = checksQuery.Where("checked_at <= ?", *to)
	}

	// Get paginated results
	checks, err := checksQuery.
		Order("checked_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(ctx)

	if err != nil {
		return nil, 0, err
	}

	return checks, total, nil
}

func (r *httpMonitorCheckRepository) Create(ctx context.Context, check model.HTTPMonitorCheckModel) (model.HTTPMonitorCheckModel, error) {
	ctx, otelSpan := trace.StartSpan(ctx, "HTTPMonitorCheckRepository.Create")
	defer otelSpan.End()

	err := gorm.G[model.HTTPMonitorCheckModel](r.DB).Create(ctx, &check)
	return check, err
}
