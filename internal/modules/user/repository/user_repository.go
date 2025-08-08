package repository

import (
	"context"
	"errors"

	"github.com/cristiano-pacheco/pingo/internal/modules/user/repository/model"
	"github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (model.UserModel, error)
	FindByEmail(ctx context.Context, email string) (model.UserModel, error)
	Create(ctx context.Context, user model.UserModel) (model.UserModel, error)
	Update(ctx context.Context, user model.UserModel) error
	IsUserActivated(ctx context.Context, userID uint64) (bool, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (model.UserModel, error) {
	user, err := gorm.G[model.UserModel](r.db).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.UserModel{}, errs.ErrRecordNotFound
		}
		return model.UserModel{}, err
	}
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (model.UserModel, error) {
	user, err := gorm.G[model.UserModel](r.db).Where("email = ?", email).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.UserModel{}, errs.ErrRecordNotFound
		}
		return model.UserModel{}, err
	}
	return user, nil
}

func (r *userRepository) Create(ctx context.Context, user model.UserModel) (model.UserModel, error) {
	err := gorm.G[model.UserModel](r.db).Create(ctx, &user)
	return user, err
}

func (r *userRepository) Update(ctx context.Context, user model.UserModel) error {
	rowsAffected, err := gorm.G[model.UserModel](r.db).Where("id = ?", user.ID).Updates(ctx, user)
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
	err := r.db.WithContext(ctx).
		Table("users").
		Select("is_activated").
		Where("id = ?", userID).
		Scan(&isActivated).Error

	if err != nil {
		return false, err
	}

	return isActivated, nil
}
