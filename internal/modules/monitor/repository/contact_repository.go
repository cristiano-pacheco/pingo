package repository

import (
	"context"

	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/model"
	"github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/database"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"gorm.io/gorm"
)

type ContactRepository interface {
	FindAll(ctx context.Context) ([]model.ContactModel, error)
	FindByName(ctx context.Context, name string) (model.ContactModel, error)
	Create(ctx context.Context, contact model.ContactModel) (model.ContactModel, error)
	Update(ctx context.Context, contact model.ContactModel) (model.ContactModel, error)
	Delete(ctx context.Context, contactID uint64) error
}

type contactRepository struct {
	*database.PingoDB
	otel otel.Otel
}

func NewContactRepository(db *database.PingoDB, otel otel.Otel) ContactRepository {
	return &contactRepository{db, otel}
}

func (r *contactRepository) FindAll(ctx context.Context) ([]model.ContactModel, error) {
	ctx, otelSpan := r.otel.StartSpan(ctx, "ContactRepository.FindAll")
	defer otelSpan.End()

	contacts, err := gorm.G[model.ContactModel](r.DB).Find(ctx)
	if err != nil {
		return nil, err
	}
	return contacts, nil
}

func (r *contactRepository) FindByName(ctx context.Context, name string) (model.ContactModel, error) {
	ctx, otelSpan := r.otel.StartSpan(ctx, "ContactRepository.FindByName")
	defer otelSpan.End()

	contact, err := gorm.G[model.ContactModel](r.DB).
		Where("name = ?", name).
		First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.ContactModel{}, errs.ErrRecordNotFound
		}
		return model.ContactModel{}, err
	}
	return contact, nil
}

func (r *contactRepository) Create(ctx context.Context, contact model.ContactModel) (model.ContactModel, error) {
	ctx, otelSpan := r.otel.StartSpan(ctx, "ContactRepository.Create")
	defer otelSpan.End()

	err := gorm.G[model.ContactModel](r.DB).Create(ctx, &contact)
	return contact, err
}

func (r *contactRepository) Update(ctx context.Context, contact model.ContactModel) (model.ContactModel, error) {
	ctx, otelSpan := r.otel.StartSpan(ctx, "ContactRepository.Update")
	defer otelSpan.End()

	_, err := gorm.G[model.ContactModel](r.DB).Updates(ctx, contact)
	if err != nil {
		return model.ContactModel{}, err
	}
	return contact, nil
}

func (r *contactRepository) Delete(ctx context.Context, contactID uint64) error {
	ctx, otelSpan := r.otel.StartSpan(ctx, "ContactRepository.Delete")
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
