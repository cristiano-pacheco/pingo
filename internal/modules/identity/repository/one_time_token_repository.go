package repository

import (
	"context"
	"errors"
	"time"

	"github.com/cristiano-pacheco/go-otel/trace"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/database"
	"gorm.io/gorm"
)

type OneTimeTokenRepository interface {
	Find(ctx context.Context, userID uint64, tokenTypeEnum enum.TokenTypeEnum) (model.OneTimeTokenModel, error)
	Create(ctx context.Context, token model.OneTimeTokenModel) (model.OneTimeTokenModel, error)
	Delete(ctx context.Context, userID uint64, tokenTypeEnum enum.TokenTypeEnum) error
}

type oneTimeTokenRepository struct {
	*database.PingoDB
}

func NewOneTimeTokenRepository(db *database.PingoDB) OneTimeTokenRepository {
	return &oneTimeTokenRepository{db}
}

func (r *oneTimeTokenRepository) Find(
	ctx context.Context,
	userID uint64,
	tokenTypeEnum enum.TokenTypeEnum,
) (model.OneTimeTokenModel, error) {
	ctx, otelSpan := trace.StartSpan(ctx, "OneTimeTokenRepository.Find")
	defer otelSpan.End()

	now := time.Now()
	token, err := gorm.G[model.OneTimeTokenModel](r.DB).
		Where("user_id = ?", userID).
		Where("token_type = ?", tokenTypeEnum.String()).
		Where("expires_at > ?", now).
		Limit(1).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.OneTimeTokenModel{}, errs.ErrRecordNotFound
		}
		return model.OneTimeTokenModel{}, err
	}
	return token, nil
}

func (r *oneTimeTokenRepository) Create(
	ctx context.Context,
	token model.OneTimeTokenModel,
) (model.OneTimeTokenModel, error) {
	ctx, otelSpan := trace.StartSpan(ctx, "OneTimeTokenRepository.Create")
	defer otelSpan.End()

	err := gorm.G[model.OneTimeTokenModel](r.DB).Create(ctx, &token)
	return token, err
}

func (r *oneTimeTokenRepository) Delete(ctx context.Context, userID uint64, tokenTypeEnum enum.TokenTypeEnum) error {
	ctx, otelSpan := trace.StartSpan(ctx, "OneTimeTokenRepository.Delete")
	defer otelSpan.End()

	rowsAffected, err := gorm.G[model.OneTimeTokenModel](r.DB).
		Where("user_id = ?", userID).
		Where("token_type = ?", tokenTypeEnum.String()).
		Delete(ctx)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errs.ErrRecordNotFound
	}
	return nil
}
