package service_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/suite"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	registry_mocks "github.com/cristiano-pacheco/pingo/internal/shared/modules/registry/mocks"
)

type TokenServiceTestSuite struct {
	suite.Suite
	sut                    service.TokenService
	privateKeyRegistryMock *registry_mocks.MockPrivateKeyRegistry
	logger                 logger.Logger
	cfg                    config.Config
	otel                   otel.Otel
	privateKey             *rsa.PrivateKey
}

func (s *TokenServiceTestSuite) SetupTest() {
	s.cfg = config.Config{
		JWT: config.JWT{
			Issuer:              "test-issuer",
			ExpirationInSeconds: 3600,
		},
		App: config.App{
			Name:    "Test App",
			Version: "1.0.0",
		},
		OpenTelemetry: config.OpenTelemetry{
			Enabled: false,
		},
		Log: config.Log{
			LogLevel: "disabled",
		},
	}

	// Create a simple no-op otel implementation for testing
	s.otel = otel.NewNoopOtel()

	s.logger = logger.New(s.cfg)

	s.privateKeyRegistryMock = registry_mocks.NewMockPrivateKeyRegistry(s.T())

	var err error
	s.privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	s.Require().NoError(err)

	s.sut = service.NewTokenService(s.cfg, s.privateKeyRegistryMock, s.logger, s.otel)
}

func TestTokenServiceSuite(t *testing.T) {
	suite.Run(t, new(TokenServiceTestSuite))
}

func (s *TokenServiceTestSuite) TestGenerateJWT_ValidUser_ReturnsValidToken() {
	// Arrange
	ctx := context.Background()
	user := model.UserModel{
		ID:        12345,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Status:    "active",
	}

	s.privateKeyRegistryMock.On("Get").Return(s.privateKey)

	// Act
	token, err := s.sut.GenerateJWT(ctx, user)

	// Assert
	s.Require().NoError(err)
	s.NotEmpty(token)

	parsedToken, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return &s.privateKey.PublicKey, nil
	})
	s.Require().NoError(err)
	s.True(parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	s.True(ok)
	s.Equal("test-issuer", claims["iss"])
	s.Equal(strconv.FormatUint(user.ID, 10), claims["sub"])
	s.NotNil(claims["exp"])
	s.NotNil(claims["iat"])
	s.NotNil(claims["nbf"])
}

func (s *TokenServiceTestSuite) TestGenerateJWT_UserWithZeroID_ReturnsValidTokenWithZeroSubject() {
	// Arrange
	ctx := context.Background()
	user := model.UserModel{
		ID:        0,
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane.doe@example.com",
		Status:    "active",
	}

	s.privateKeyRegistryMock.On("Get").Return(s.privateKey)

	// Act
	token, err := s.sut.GenerateJWT(ctx, user)

	// Assert
	s.Require().NoError(err)
	s.NotEmpty(token)

	parsedToken, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return &s.privateKey.PublicKey, nil
	})
	s.Require().NoError(err)
	s.True(parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	s.True(ok)
	s.Equal("0", claims["sub"])
}

func (s *TokenServiceTestSuite) TestGenerateJWT_PrivateKeySigningFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	user := model.UserModel{
		ID:        12345,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Status:    "active",
	}

	invalidKey := &rsa.PrivateKey{}
	s.privateKeyRegistryMock.On("Get").Return(invalidKey)

	// Act
	token, err := s.sut.GenerateJWT(ctx, user)

	// Assert
	s.Require().Error(err)
	s.Empty(token)
}

func (s *TokenServiceTestSuite) TestGenerateJWT_TokenExpirationIsSetCorrectly_ReturnsValidToken() {
	// Arrange
	ctx := context.Background()
	user := model.UserModel{
		ID:        12345,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Status:    "active",
	}

	s.privateKeyRegistryMock.On("Get").Return(s.privateKey)

	beforeGeneration := time.Now()

	// Act
	token, err := s.sut.GenerateJWT(ctx, user)

	// Assert
	s.Require().NoError(err)
	s.NotEmpty(token)

	afterGeneration := time.Now()

	parsedToken, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return &s.privateKey.PublicKey, nil
	})
	s.Require().NoError(err)
	s.True(parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	s.True(ok)

	expClaim, ok := claims["exp"].(float64)
	s.True(ok)
	expTime := time.Unix(int64(expClaim), 0)

	expectedMinExpiration := beforeGeneration.Add(time.Duration(s.cfg.JWT.ExpirationInSeconds) * time.Second)
	expectedMaxExpiration := afterGeneration.Add(time.Duration(s.cfg.JWT.ExpirationInSeconds) * time.Second)

	s.True(expTime.After(expectedMinExpiration.Add(-time.Second)) || expTime.Equal(expectedMinExpiration))
	s.True(expTime.Before(expectedMaxExpiration.Add(time.Second)) || expTime.Equal(expectedMaxExpiration))
}

func (s *TokenServiceTestSuite) TestGenerateJWT_AllClaimsAreSetCorrectly_ReturnsValidToken() {
	// Arrange
	ctx := context.Background()
	user := model.UserModel{
		ID:        99999,
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice.smith@example.com",
		Status:    "verified",
	}

	s.privateKeyRegistryMock.On("Get").Return(s.privateKey)

	beforeGeneration := time.Now()

	// Act
	token, err := s.sut.GenerateJWT(ctx, user)

	// Assert
	s.Require().NoError(err)
	s.NotEmpty(token)

	afterGeneration := time.Now()

	parsedToken, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return &s.privateKey.PublicKey, nil
	})
	s.Require().NoError(err)
	s.True(parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	s.True(ok)

	s.Equal(s.cfg.JWT.Issuer, claims["iss"])
	s.Equal(strconv.FormatUint(user.ID, 10), claims["sub"])

	iatClaim, ok := claims["iat"].(float64)
	s.True(ok)
	iatTime := time.Unix(int64(iatClaim), 0)
	s.True(iatTime.After(beforeGeneration.Add(-time.Second)) || iatTime.Equal(beforeGeneration))
	s.True(iatTime.Before(afterGeneration.Add(time.Second)) || iatTime.Equal(afterGeneration))

	nbfClaim, ok := claims["nbf"].(float64)
	s.True(ok)
	nbfTime := time.Unix(int64(nbfClaim), 0)
	s.True(nbfTime.After(beforeGeneration.Add(-time.Second)) || nbfTime.Equal(beforeGeneration))
	s.True(nbfTime.Before(afterGeneration.Add(time.Second)) || nbfTime.Equal(afterGeneration))
}
