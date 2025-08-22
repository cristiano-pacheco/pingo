package repository

import (
	"context"
	"errors"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/database"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(ctx context.Context, userID uint64) (model.UserModel, error)
	FindByEmail(ctx context.Context, email string) (model.UserModel, error)
	FindPendingConfirmation(ctx context.Context, confirmationToken []byte) (model.UserModel, error)
	Create(ctx context.Context, user model.UserModel) (model.UserModel, error)
	Update(ctx context.Context, user model.UserModel) error
	IsUserActivated(ctx context.Context, userID uint64) (bool, error)
}

type userRepository struct {
	*database.PingoDB
}

func NewUserRepository(db *database.PingoDB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) FindByID(ctx context.Context, userID uint64) (model.UserModel, error) {
	user, err := gorm.G[model.UserModel](r.DB).Limit(1).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.UserModel{}, errs.ErrRecordNotFound
		}
		return model.UserModel{}, err
	}
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (model.UserModel, error) {
	user, err := gorm.G[model.UserModel](r.DB).Where("email = ?", email).Limit(1).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.UserModel{}, errs.ErrRecordNotFound
		}
		return model.UserModel{}, err
	}
	return user, nil
}

func (r *userRepository) FindPendingConfirmation(
	ctx context.Context,
	confirmationToken []byte,
) (model.UserModel, error) {
	user, err := gorm.G[model.UserModel](r.DB).
		Where("confirmation_token = ?", confirmationToken).
		Where("status = ?", "pending").
		Limit(1).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.UserModel{}, errs.ErrRecordNotFound
		}
		return model.UserModel{}, err
	}
	return user, nil
}

func (r *userRepository) Create(ctx context.Context, user model.UserModel) (model.UserModel, error) {
	err := gorm.G[model.UserModel](r.DB).Create(ctx, &user)
	return user, err
}

func (r *userRepository) Update(ctx context.Context, user model.UserModel) error {
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
