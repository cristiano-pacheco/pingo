package repository

import (
	"context"
	"errors"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/model"
	"github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/database"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(ctx context.Context, userID uint64) (model.UserModel, error)
	FindByEmail(ctx context.Context, email string) (model.UserModel, error)
	FindByConfirmationToken(ctx context.Context, confirmationToken []byte) (model.UserModel, error)
	Create(ctx context.Context, user model.UserModel) (model.UserModel, error)
	Update(ctx context.Context, user model.UserModel) error
	IsUserActivated(ctx context.Context, userID uint64) (bool, error)
}

type userRepository struct {
	db *database.PingoDB
}

func NewUserRepository(db *database.PingoDB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, userID uint64) (model.UserModel, error) {
	user, err := gorm.G[model.UserModel](r.db.DB).Limit(1).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.UserModel{}, errs.ErrRecordNotFound
		}
		return model.UserModel{}, err
	}
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (model.UserModel, error) {
	user, err := gorm.G[model.UserModel](r.db.DB).Where("email = ?", email).Limit(1).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.UserModel{}, errs.ErrRecordNotFound
		}
		return model.UserModel{}, err
	}
	return user, nil
}

func (r *userRepository) FindByConfirmationToken(ctx context.Context, confirmationToken []byte) (model.UserModel, error) {
	user, err := gorm.G[model.UserModel](r.db.DB).
		Where("confirmation_token = ?", confirmationToken).Limit(1).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.UserModel{}, errs.ErrRecordNotFound
		}
		return model.UserModel{}, err
	}
	return user, nil
}

func (r *userRepository) Create(ctx context.Context, user model.UserModel) (model.UserModel, error) {
	err := gorm.G[model.UserModel](r.db.DB).Create(ctx, &user)
	return user, err
}

func (r *userRepository) Update(ctx context.Context, user model.UserModel) error {
	rowsAffected, err := gorm.G[model.UserModel](r.db.DB).Where("id = ?", user.ID).Updates(ctx, user)
	if rowsAffected == 0 {
		return errs.ErrRecordNotFound
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) IsUserActivated(ctx context.Context, userID uint64) (bool, error) {
	var isActivated bool
	err := r.db.DB.WithContext(ctx).
		Table("users").
		Select("is_activated").
		Where("id = ?", userID).
		Scan(&isActivated).Error

	if err != nil {
		return false, err
	}

	return isActivated, nil
}
