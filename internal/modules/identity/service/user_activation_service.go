package service

import (
	"context"
	"errors"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/cache"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
)

type UserActivationService interface {
	IsUserActivated(ctx context.Context, userID uint64) (bool, error)
}

type userActivationService struct {
	userActivatedCache cache.UserActivatedCache
	userRepository     repository.UserRepository
	logger             logger.Logger
}

func NewUserActivationService(
	userActivatedCache cache.UserActivatedCache,
	userRepository repository.UserRepository,
	logger logger.Logger,
) UserActivationService {
	return &userActivationService{
		userActivatedCache: userActivatedCache,
		userRepository:     userRepository,
		logger:             logger,
	}
}

func (s *userActivationService) IsUserActivated(ctx context.Context, userID uint64) (bool, error) {
	// Try cache first for fast lookup
	isActivated, err := s.userActivatedCache.Get(userID)
	if err != nil {
		// Log cache error but continue with database lookup
		s.logger.Warn().Msgf("Failed to check user activation cache for user_id: %d, error: %v", userID, err)
	} else if isActivated {
		// Cache hit - user is activated
		return true, nil
	}

	// Cache miss or error - check database
	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, shared_errs.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	isActive := user.Status == enum.UserStatusActive

	// If user is active, update cache for future lookups
	if isActive {
		cacheErr := s.userActivatedCache.Set(userID)
		if cacheErr != nil {
			// Log cache error but don't fail the request
			s.logger.Warn().Msgf("Failed to set user activation cache for user_id: %d, error: %v", userID, cacheErr)
		}
	}

	return isActive, nil
}
