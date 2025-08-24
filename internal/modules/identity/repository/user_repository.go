package repository

import (
	"context"
	"errors"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/database"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(ctx context.Context, userID uint64) (model.UserModel, error)
	FindByEmail(ctx context.Context, email string) (model.UserModel, error)
	Create(ctx context.Context, user model.UserModel) (model.UserModel, error)
	Update(ctx context.Context, user model.UserModel) error
	IsUserActivated(ctx context.Context, userID uint64) (bool, error)
}

type userRepository struct {
	*database.PingoDB
	otel otel.Otel
}

func NewUserRepository(db *database.PingoDB, otel otel.Otel) UserRepository {
	return &userRepository{db, otel}
}

func (r *userRepository) FindByID(ctx context.Context, userID uint64) (model.UserModel, error) {
	ctx, otelSpan := r.otel.StartSpan(ctx, "UserRepository.FindByID")
	defer otelSpan.End()

	user, err := gorm.G[model.UserModel](r.DB).Limit(1).Where("id = ?", userID).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.UserModel{}, errs.ErrRecordNotFound
		}
		return model.UserModel{}, err
	}
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (model.UserModel, error) {
	ctx, otelSpan := r.otel.StartSpan(ctx, "UserRepository.FindByEmail")
	defer otelSpan.End()

	user, err := gorm.G[model.UserModel](r.DB).Where("email = ?", email).Limit(1).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.UserModel{}, errs.ErrRecordNotFound
		}
		return model.UserModel{}, err
	}
	return user, nil
}

func (r *userRepository) Create(ctx context.Context, user model.UserModel) (model.UserModel, error) {
	ctx, otelSpan := r.otel.StartSpan(ctx, "UserRepository.Create")
	defer otelSpan.End()

	err := gorm.G[model.UserModel](r.DB).Create(ctx, &user)
	return user, err
}

func (r *userRepository) Update(ctx context.Context, user model.UserModel) error {
	ctx, otelSpan := r.otel.StartSpan(ctx, "UserRepository.Update")
	defer otelSpan.End()

	rowsAffected, err := gorm.G[model.UserModel](r.DB).Where("id = ?", user.ID).Updates(ctx, user)
	if rowsAffected == 0 {
		return errs.ErrRecordNotFound
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) IsUserActivated(ctx context.Context, userID uint64) (bool, error) {
	ctx, otelSpan := r.otel.StartSpan(ctx, "UserRepository.IsUserActivated")
	defer otelSpan.End()

	user, err := r.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}

	if user.ID == 0 {
		return false, nil
	}

	if user.Status == enum.UserStatusActive {
		return true, nil
	}

	return false, nil
}
