package service_test

import (
	"testing"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type HashServiceTestSuite struct {
	suite.Suite
	sut service.HashService
}

func (s *HashServiceTestSuite) SetupTest() {
	s.sut = service.NewHashService()
}

func TestHashServiceSuite(t *testing.T) {
	suite.Run(t, new(HashServiceTestSuite))
}

func (s *HashServiceTestSuite) TestGenerateFromPassword_ValidPassword_ReturnsHashedPassword() {
	// Arrange
	password := []byte("testpassword123")

	// Act
	hashedPassword, err := s.sut.GenerateFromPassword(password)

	// Assert
	s.NoError(err)
	s.NotNil(hashedPassword)
	s.NotEqual(password, hashedPassword)
	s.Greater(len(hashedPassword), 0)
}

func (s *HashServiceTestSuite) TestGenerateFromPassword_EmptyPassword_ReturnsHashedPassword() {
	// Arrange
	password := []byte("")

	// Act
	hashedPassword, err := s.sut.GenerateFromPassword(password)

	// Assert
	s.NoError(err)
	s.NotNil(hashedPassword)
	s.Greater(len(hashedPassword), 0)
}

func (s *HashServiceTestSuite) TestCompareHashAndPassword_ValidPasswordAndHash_ReturnsNoError() {
	// Arrange
	password := []byte("testpassword123")
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	s.NoError(err)

	// Act
	err = s.sut.CompareHashAndPassword(hashedPassword, password)

	// Assert
	s.NoError(err)
}

func (s *HashServiceTestSuite) TestCompareHashAndPassword_InvalidPassword_ReturnsError() {
	// Arrange
	password := []byte("testpassword123")
	wrongPassword := []byte("wrongpassword")
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	s.NoError(err)

	// Act
	err = s.sut.CompareHashAndPassword(hashedPassword, wrongPassword)

	// Assert
	s.Error(err)
	s.ErrorIs(err, bcrypt.ErrMismatchedHashAndPassword)
}

func (s *HashServiceTestSuite) TestCompareHashAndPassword_InvalidHash_ReturnsError() {
	// Arrange
	password := []byte("testpassword123")
	invalidHash := []byte("invalidhash")

	// Act
	err := s.sut.CompareHashAndPassword(invalidHash, password)

	// Assert
	s.Error(err)
}

func (s *HashServiceTestSuite) TestGenerateRandomBytes_DefaultSize_ReturnsRandomBytes() {
	// Arrange
	expectedSize := 128

	// Act
	randomBytes, err := s.sut.GenerateRandomBytes()

	// Assert
	s.NoError(err)
	s.NotNil(randomBytes)
	s.Equal(expectedSize, len(randomBytes))
}

func (s *HashServiceTestSuite) TestGenerateRandomBytes_MultipleCalls_ReturnsDifferentBytes() {
	// Arrange

	// Act
	randomBytes1, err1 := s.sut.GenerateRandomBytes()
	randomBytes2, err2 := s.sut.GenerateRandomBytes()

	// Assert
	s.NoError(err1)
	s.NoError(err2)
	s.NotEqual(randomBytes1, randomBytes2)
}
