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

type HTTPMonitorRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]model.HTTPMonitorModel, int64, error)
	FindByID(ctx context.Context, monitorID uint64) (model.HTTPMonitorModel, error)
	Create(ctx context.Context, monitor model.HTTPMonitorModel) (model.HTTPMonitorModel, error)
	Update(ctx context.Context, monitor model.HTTPMonitorModel) (model.HTTPMonitorModel, error)
	Delete(ctx context.Context, monitorID uint64) error
	AssignContacts(ctx context.Context, monitorID uint64, contactIDs []uint64) error
}

type httpMonitorRepository struct {
	*database.PingoDB
}

func NewHTTPMonitorRepository(db *database.PingoDB) HTTPMonitorRepository {
	return &httpMonitorRepository{db}
}

func (r *httpMonitorRepository) FindAll(ctx context.Context, page, pageSize int) ([]model.HTTPMonitorModel, int64, error) {
	ctx, otelSpan := trace.StartSpan(ctx, "HTTPMonitorRepository.FindAll")
	defer otelSpan.End()

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	if err := r.DB.Model(&model.HTTPMonitorModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	monitors, err := gorm.G[model.HTTPMonitorModel](r.DB).
		Limit(pageSize).
		Offset(offset).
		Find(ctx)
	if err != nil {
		return nil, 0, err
	}

	return monitors, total, nil
}

func (r *httpMonitorRepository) FindByID(ctx context.Context, monitorID uint64) (model.HTTPMonitorModel, error) {
	ctx, otelSpan := trace.StartSpan(ctx, "HTTPMonitorRepository.FindByID")
	defer otelSpan.End()

	monitor, err := gorm.G[model.HTTPMonitorModel](r.DB).
		Where("id = ?", monitorID).
		Limit(1).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.HTTPMonitorModel{}, errs.ErrRecordNotFound
		}
		return model.HTTPMonitorModel{}, err
	}
	return monitor, nil
}

func (r *httpMonitorRepository) Create(ctx context.Context, monitor model.HTTPMonitorModel) (model.HTTPMonitorModel, error) {
	ctx, otelSpan := trace.StartSpan(ctx, "HTTPMonitorRepository.Create")
	defer otelSpan.End()

	err := gorm.G[model.HTTPMonitorModel](r.DB).Create(ctx, &monitor)
	return monitor, err
}

func (r *httpMonitorRepository) Update(ctx context.Context, monitor model.HTTPMonitorModel) (model.HTTPMonitorModel, error) {
	ctx, otelSpan := trace.StartSpan(ctx, "HTTPMonitorRepository.Update")
	defer otelSpan.End()

	_, err := gorm.G[model.HTTPMonitorModel](r.DB).Updates(ctx, monitor)
	if err != nil {
		return model.HTTPMonitorModel{}, err
	}
	return monitor, nil
}

func (r *httpMonitorRepository) Delete(ctx context.Context, monitorID uint64) error {
	ctx, otelSpan := trace.StartSpan(ctx, "HTTPMonitorRepository.Delete")
	defer otelSpan.End()

	rowsAffected, err := gorm.G[model.HTTPMonitorModel](r.DB).
		Where("id = ?", monitorID).
		Delete(ctx)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errs.ErrRecordNotFound
	}
	return nil
}

func (r *httpMonitorRepository) AssignContacts(ctx context.Context, monitorID uint64, contactIDs []uint64) error {
	ctx, otelSpan := trace.StartSpan(ctx, "HTTPMonitorRepository.AssignContacts")
	defer otelSpan.End()

	// start a transaction
	tx := r.DB.Begin()

	_, err := gorm.G[model.HTTPMonitorContactModel](tx).
		Where("http_monitor_id = ?", monitorID).
		Delete(ctx)

	if err != nil {
		tx.Rollback()
		return err
	}

	var monitorContacts []model.HTTPMonitorContactModel
	for _, contactID := range contactIDs {
		monitorContacts = append(monitorContacts, model.HTTPMonitorContactModel{
			HTTPMonitorID: monitorID,
			ContactID:     contactID,
		})
	}

	err = gorm.G[model.HTTPMonitorContactModel](tx).CreateInBatches(ctx, &monitorContacts, len(monitorContacts))
	if err != nil {
		tx.Rollback()
		return err
	}

	if commitErr := tx.Commit().Error; commitErr != nil {
		return commitErr
	}

	return nil
}
