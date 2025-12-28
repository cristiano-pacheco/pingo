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

type OneTimeTokenRepositoryI interface {
	Find(ctx context.Context, userID uint64, tokenTypeEnum enum.TokenTypeEnum) (model.OneTimeTokenModel, error)
	Create(ctx context.Context, token model.OneTimeTokenModel) (model.OneTimeTokenModel, error)
	Delete(ctx context.Context, userID uint64, tokenTypeEnum enum.TokenTypeEnum) error
}

type OneTimeTokenRepository struct {
	*database.PingoDB
}

var _ OneTimeTokenRepositoryI = (*OneTimeTokenRepository)(nil)

func NewOneTimeTokenRepository(db *database.PingoDB) *OneTimeTokenRepository {
	return &OneTimeTokenRepository{db}
}

func (r *OneTimeTokenRepository) Find(
	ctx context.Context,
	userID uint64,
	tokenTypeEnum enum.TokenTypeEnum,
) (model.OneTimeTokenModel, error) {
	ctx, otelSpan := trace.Span(ctx, "OneTimeTokenRepository.Find")
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

func (r *OneTimeTokenRepository) Create(
	ctx context.Context,
	token model.OneTimeTokenModel,
) (model.OneTimeTokenModel, error) {
	ctx, otelSpan := trace.Span(ctx, "OneTimeTokenRepository.Create")
	defer otelSpan.End()

	err := gorm.G[model.OneTimeTokenModel](r.DB).Create(ctx, &token)
	return token, err
}

func (r *OneTimeTokenRepository) Delete(ctx context.Context, userID uint64, tokenTypeEnum enum.TokenTypeEnum) error {
	ctx, otelSpan := trace.Span(ctx, "OneTimeTokenRepository.Delete")
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
