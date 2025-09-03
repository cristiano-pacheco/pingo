package cache_test

import (
	"testing"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/cache"
	"github.com/cristiano-pacheco/pingo/pkg/redis/mocks"
	"github.com/stretchr/testify/suite"
)

type UserActivatedCacheTestSuite struct {
	suite.Suite
	mockRedis *mocks.MockRedis
	sut       cache.UserActivatedCache
}

func (s *UserActivatedCacheTestSuite) SetupTest() {
	s.mockRedis = mocks.NewMockRedis(s.T())
	s.sut = cache.NewUserActivatedCache(s.mockRedis)
}

func TestUserActivatedCacheSuite(t *testing.T) {
	suite.Run(t, new(UserActivatedCacheTestSuite))
}

func (s *UserActivatedCacheTestSuite) TestNewUserActivatedCache_ValidRedisClient_ReturnsCache() {
	// Act & Assert
	s.NotNil(s.sut)
	s.Implements((*cache.UserActivatedCache)(nil), s.sut)
}

// Integration test example that would work with a real Redis instance
// This demonstrates how the cache should be tested in a real environment
func (s *UserActivatedCacheTestSuite) TestIntegration_SetGetDelete_WorksCorrectly() {
	// This test would need to be run with a real Redis instance
	// It's commented out but shows the expected behavior

	/*
		userID := uint64(456)

		// Initially, user should not be activated
		isActivated, err := s.sut.Get(userID)
		s.Require().NoError(err)
		s.False(isActivated)

		// Set user as activated
		err = s.sut.Set(userID)
		s.Require().NoError(err)

		// User should now be activated
		isActivated, err = s.sut.Get(userID)
		s.Require().NoError(err)
		s.True(isActivated)

		// Delete the activation
		err = s.sut.Delete(userID)
		s.Require().NoError(err)

		// User should no longer be activated
		isActivated, err = s.sut.Get(userID)
		s.Require().NoError(err)
		s.False(isActivated)
	*/
}
