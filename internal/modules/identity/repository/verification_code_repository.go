package repository

import (
	"context"
	"errors"
	"time"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/database"
	"gorm.io/gorm"
)

type VerificationCodeRepository interface {
	FindByUserAndCode(ctx context.Context, userID uint, code string) (model.VerificationCodeModel, error)
	Create(ctx context.Context, v model.VerificationCodeModel) (model.VerificationCodeModel, error)
	Update(ctx context.Context, v model.VerificationCodeModel) error
	DeleteByUserID(ctx context.Context, userID uint64) error
}

type verificationCodeRepository struct {
	*database.PingoDB
}

func NewVerificationCodeRepository(db *database.PingoDB) VerificationCodeRepository {
	return &verificationCodeRepository{db}
}

func (r *verificationCodeRepository) FindByUserAndCode(ctx context.Context, userID uint, code string) (model.VerificationCodeModel, error) {
	now := time.Now()
	v, err := gorm.G[model.VerificationCodeModel](r.DB).
		Where("user_id = ?", userID).
		Where("code = ?", code).
		Where("used_at IS NULL").
		Where("expires_at > ?", now).
		Limit(1).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.VerificationCodeModel{}, errs.ErrRecordNotFound
		}
		return model.VerificationCodeModel{}, err
	}
	return v, nil
}

func (r *verificationCodeRepository) Create(ctx context.Context, v model.VerificationCodeModel) (model.VerificationCodeModel, error) {
	err := gorm.G[model.VerificationCodeModel](r.DB).Create(ctx, &v)
	return v, err
}

func (r *verificationCodeRepository) Update(ctx context.Context, v model.VerificationCodeModel) error {
	rowsAffected, err := gorm.G[model.VerificationCodeModel](r.DB).Where("id = ?", v.ID).Updates(ctx, v)
	if rowsAffected == 0 {
		return errs.ErrRecordNotFound
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *verificationCodeRepository) DeleteByUserID(ctx context.Context, userID uint64) error {
	rowsAffected, err := gorm.G[model.VerificationCodeModel](r.DB).Where("user_id = ?", userID).Delete(ctx)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errs.ErrRecordNotFound
	}
	return nil
}
