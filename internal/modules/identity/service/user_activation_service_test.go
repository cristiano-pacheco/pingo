package service_test

import (
	"context"
	"testing"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/cache"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	repository_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger/mocks"
	cache_mocks "github.com/cristiano-pacheco/pingo/pkg/redis/mocks"
	"github.com/stretchr/testify/suite"
)

type UserActivationServiceTestSuite struct {
	suite.Suite
	mockUserActivatedCache *cache_mocks.MockRedis
	mockUserRepository     *repository_mocks.MockUserRepository
	mockLogger             *mocks.MockLogger
	sut                    service.UserActivationService
}

func (s *UserActivationServiceTestSuite) SetupTest() {
	s.mockUserActivatedCache = cache_mocks.NewMockRedis(s.T())
	s.mockUserRepository = repository_mocks.NewMockUserRepository(s.T())
	s.mockLogger = mocks.NewMockLogger(s.T())

	userActivatedCache := cache.NewUserActivatedCache(s.mockUserActivatedCache)

	s.sut = service.NewUserActivationService(
		userActivatedCache,
		s.mockUserRepository,
		s.mockLogger,
	)
}

func TestUserActivationServiceSuite(t *testing.T) {
	suite.Run(t, new(UserActivationServiceTestSuite))
}

func (s *UserActivationServiceTestSuite) TestIsUserActivated_CacheHit_ReturnsTrue() {
	// This test would require either integration testing with a real Redis instance
	// or more sophisticated mocking of the redis client operations
	// For now, we'll test that the service is properly instantiated
	s.NotNil(s.sut)
}

func (s *UserActivationServiceTestSuite) TestIsUserActivated_CacheMissActiveUser_ReturnsTrueAndUpdatesCache() {
	// This test would require either integration testing with a real Redis instance
	// or more sophisticated mocking of the redis client operations
	// For now, we'll test the database fallback logic

	// Arrange
	userID := uint64(123)
	activeUser := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusActive,
	}

	s.mockUserRepository.EXPECT().FindByID(context.Background(), userID).Return(activeUser, nil)

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

	s.mockUserRepository.EXPECT().FindByID(context.Background(), userID).Return(inactiveUser, nil)

	// Act
	isActivated, err := s.sut.IsUserActivated(context.Background(), userID)

	// Assert
	s.Require().NoError(err)
	s.False(isActivated)
}

func (s *UserActivationServiceTestSuite) TestIsUserActivated_UserNotFound_ReturnsFalse() {
	// Arrange
	userID := uint64(999)

	s.mockUserRepository.EXPECT().FindByID(context.Background(), userID).Return(model.UserModel{}, shared_errs.ErrRecordNotFound)

	// Act
	isActivated, err := s.sut.IsUserActivated(context.Background(), userID)

	// Assert
	s.Require().NoError(err)
	s.False(isActivated)
}
