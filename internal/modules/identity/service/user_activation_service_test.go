package service_test

import (
	"context"
	"errors"
	"testing"

	cache_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/cache/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	repository_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger/mocks"
	"github.com/stretchr/testify/suite"
)

type UserActivationServiceTestSuite struct {
	suite.Suite
	mockUserActivatedCache *cache_mocks.MockUserActivatedCacheI
	mockUserRepository     *repository_mocks.MockUserRepositoryI
	mockLogger             *mocks.MockLogger
	sut                    *service.UserActivationService
}

func (s *UserActivationServiceTestSuite) SetupTest() {
	s.mockUserActivatedCache = cache_mocks.NewMockUserActivatedCacheI(s.T())
	s.mockUserRepository = repository_mocks.NewMockUserRepositoryI(s.T())
	s.mockLogger = mocks.NewMockLogger(s.T())

	s.sut = service.NewUserActivationService(
		s.mockUserActivatedCache,
		s.mockUserRepository,
		s.mockLogger,
	)
}

func TestUserActivationServiceSuite(t *testing.T) {
	suite.Run(t, new(UserActivationServiceTestSuite))
}

func (s *UserActivationServiceTestSuite) TestIsUserActivated_CacheHit_ReturnsTrue() {
	// Arrange
	userID := uint64(123)

	// Mock cache hit - user is activated and found in cache
	s.mockUserActivatedCache.EXPECT().Get(userID).Return(true, nil).Once()
	// Repository should not be called when cache hits

	// Act
	isActivated, err := s.sut.IsUserActivated(context.Background(), userID)

	// Assert
	s.Require().NoError(err)
	s.True(isActivated)
}

func (s *UserActivationServiceTestSuite) TestIsUserActivated_CacheMissActiveUser_ReturnsTrueAndUpdatesCache() {
	// Arrange
	userID := uint64(123)
	activeUser := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusActive,
	}

	// Mock cache operations - first call returns false (cache miss), second call sets cache
	s.mockUserActivatedCache.EXPECT().Get(userID).Return(false, nil).Once()
	s.mockUserRepository.EXPECT().FindByID(context.Background(), userID).Return(activeUser, nil).Once()
	s.mockUserActivatedCache.EXPECT().Set(userID).Return(nil).Once()

	// Act
	isActivated, err := s.sut.IsUserActivated(context.Background(), userID)

	// Assert
	s.Require().NoError(err)
	s.True(isActivated)
}

func (s *UserActivationServiceTestSuite) TestIsUserActivated_CacheMissInactiveUser_ReturnsFalse() {
	// Arrange
	userID := uint64(123)
	inactiveUser := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusPending,
	}

	// Mock cache operations - cache miss, then don't set cache for inactive user
	s.mockUserActivatedCache.EXPECT().Get(userID).Return(false, nil).Once()
	s.mockUserRepository.EXPECT().FindByID(context.Background(), userID).Return(inactiveUser, nil).Once()

	// Act
	isActivated, err := s.sut.IsUserActivated(context.Background(), userID)

	// Assert
	s.Require().NoError(err)
	s.False(isActivated)
}

func (s *UserActivationServiceTestSuite) TestIsUserActivated_UserNotFound_ReturnsFalse() {
	// Arrange
	userID := uint64(999)

	// Mock cache operations - cache miss, then user not found in database
	s.mockUserActivatedCache.EXPECT().Get(userID).Return(false, nil).Once()
	s.mockUserRepository.EXPECT().
		FindByID(context.Background(), userID).
		Return(model.UserModel{}, shared_errs.ErrRecordNotFound).Once()

	// Act
	isActivated, err := s.sut.IsUserActivated(context.Background(), userID)

	// Assert
	s.Require().NoError(err)
	s.False(isActivated)
}

func (s *UserActivationServiceTestSuite) TestIsUserActivated_CacheErrorFallsBackToDatabase() {
	// Arrange
	userID := uint64(123)
	activeUser := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusActive,
	}

	// Mock cache error, then successful database lookup
	s.mockUserActivatedCache.EXPECT().Get(userID).Return(false, errors.New("cache error")).Once()
	s.mockLogger.EXPECT().Warn().Return(nil).Once() // Mock logger warn call
	s.mockUserRepository.EXPECT().FindByID(context.Background(), userID).Return(activeUser, nil).Once()
	s.mockUserActivatedCache.EXPECT().Set(userID).Return(nil).Once()

	// Act
	isActivated, err := s.sut.IsUserActivated(context.Background(), userID)

	// Assert
	s.Require().NoError(err)
	s.True(isActivated)
}

func (s *UserActivationServiceTestSuite) TestIsUserActivated_CacheSetError_DoesNotFailRequest() {
	// Arrange
	userID := uint64(123)
	activeUser := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusActive,
	}

	// Mock cache miss, successful database lookup, but cache set fails
	s.mockUserActivatedCache.EXPECT().Get(userID).Return(false, nil).Once()
	s.mockUserRepository.EXPECT().FindByID(context.Background(), userID).Return(activeUser, nil).Once()
	s.mockUserActivatedCache.EXPECT().Set(userID).Return(errors.New("cache set error")).Once()
	s.mockLogger.EXPECT().Warn().Return(nil).Once() // Mock logger warn call

	// Act
	isActivated, err := s.sut.IsUserActivated(context.Background(), userID)

	// Assert
	s.Require().NoError(err) // Should not fail even if cache set fails
	s.True(isActivated)
}
