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

type ContactRepositoryI interface {
	FindAll(ctx context.Context) ([]model.ContactModel, error)
	FindByName(ctx context.Context, name string) (model.ContactModel, error)
	Create(ctx context.Context, contact model.ContactModel) (model.ContactModel, error)
	Update(ctx context.Context, contact model.ContactModel) (model.ContactModel, error)
	Delete(ctx context.Context, contactID uint64) error
}

type ContactRepository struct {
	*database.PingoDB
}

var _ ContactRepositoryI = (*ContactRepository)(nil)

func NewContactRepository(db *database.PingoDB) *ContactRepository {
	return &ContactRepository{db}
}

func (r *ContactRepository) FindAll(ctx context.Context) ([]model.ContactModel, error) {
	ctx, otelSpan := trace.Span(ctx, "ContactRepository.FindAll")
	defer otelSpan.End()

	contacts, err := gorm.G[model.ContactModel](r.DB).Find(ctx)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

func (r *ContactRepository) FindByName(ctx context.Context, name string) (model.ContactModel, error) {
	ctx, otelSpan := trace.Span(ctx, "ContactRepository.FindByName")
	defer otelSpan.End()

	contact, err := gorm.G[model.ContactModel](r.DB).
		Where("name = ?", name).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.ContactModel{}, errs.ErrRecordNotFound
		}
		return model.ContactModel{}, err
	}
	return contact, nil
}

func (r *ContactRepository) Create(ctx context.Context, contact model.ContactModel) (model.ContactModel, error) {
	ctx, otelSpan := trace.Span(ctx, "ContactRepository.Create")
	defer otelSpan.End()

	err := gorm.G[model.ContactModel](r.DB).Create(ctx, &contact)
	return contact, err
}

func (r *ContactRepository) Update(ctx context.Context, contact model.ContactModel) (model.ContactModel, error) {
	ctx, otelSpan := trace.Span(ctx, "ContactRepository.Update")
	defer otelSpan.End()

	_, err := gorm.G[model.ContactModel](r.DB).Updates(ctx, contact)
	if err != nil {
		return model.ContactModel{}, err
	}
	return contact, nil
}

func (r *ContactRepository) Delete(ctx context.Context, contactID uint64) error {
	ctx, otelSpan := trace.Span(ctx, "ContactRepository.Delete")
	defer otelSpan.End()

	rowsAffected, err := gorm.G[model.ContactModel](r.DB).
		Where("id = ?", contactID).
		Delete(ctx)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errs.ErrRecordNotFound
	}
	return nil
}
