package consumer_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event/consumer"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	repository_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	service_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/service/mocks"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"

	"github.com/cristiano-pacheco/pingo/pkg/kafka"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UserCreatedConsumerTestSuite struct {
	suite.Suite
	sut                          *consumer.UserCreatedConsumer
	sendEmailConfirmationService *service_mocks.MockSendEmailConfirmationServiceI
	oneTimeTokenRepository       *repository_mocks.MockOneTimeTokenRepositoryI
	userRepository               *repository_mocks.MockUserRepositoryI
	hashService                  *service_mocks.MockHashServiceI
	logger                       logger.Logger
}

func (s *UserCreatedConsumerTestSuite) SetupTest() {
	s.sendEmailConfirmationService = service_mocks.NewMockSendEmailConfirmationServiceI(s.T())
	s.oneTimeTokenRepository = repository_mocks.NewMockOneTimeTokenRepositoryI(s.T())
	s.userRepository = repository_mocks.NewMockUserRepositoryI(s.T())
	s.hashService = service_mocks.NewMockHashServiceI(s.T())

	// Use a disabled logger for testing
	loggerConfig := config.Config{
		Log: config.Log{
			LogLevel: "disabled",
		},
		OpenTelemetry: config.OpenTelemetry{
			Enabled: false,
		},
	}
	s.logger = logger.New(loggerConfig)

	s.sut = consumer.NewUserCreatedConsumer(
		s.sendEmailConfirmationService,
		s.oneTimeTokenRepository,
		s.userRepository,
		s.hashService,
		s.logger,
	)
}

func TestUserCreatedConsumerSuite(t *testing.T) {
	suite.Run(t, new(UserCreatedConsumerTestSuite))
}

func (s *UserCreatedConsumerTestSuite) TestTopic_ReturnsCorrectTopic() {
	// Act
	topic := s.sut.Topic()

	// Assert
	s.Equal(event.IdentityUserCreatedTopic, topic)
}

func (s *UserCreatedConsumerTestSuite) TestGroupID_ReturnsDefaultGroupID() {
	// Act
	groupID := s.sut.GroupID()

	// Assert
	s.Equal("default", groupID)
}

func (s *UserCreatedConsumerTestSuite) TestProcessMessage_ValidMessage_ProcessesSuccessfully() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userCreatedMessage := event.UserCreatedMessage{
		UserID: userID,
	}
	messageBytes, err := json.Marshal(userCreatedMessage)
	s.Require().NoError(err)

	message := kafka.Message{
		Value: messageBytes,
	}

	user := model.UserModel{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	tokenBytes := []byte("random-token-bytes")

	expectedOneTimeToken := model.OneTimeTokenModel{
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		TokenType: enum.TokenTypeAccountConfirmation,
		TokenHash: tokenBytes,
	}

	createdToken := model.OneTimeTokenModel{
		ID:        1,
		UserID:    user.ID,
		ExpiresAt: expectedOneTimeToken.ExpiresAt,
		TokenType: enum.TokenTypeAccountConfirmation,
		TokenHash: tokenBytes,
		CreatedAt: time.Now(),
	}

	// Mock expectations
	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.hashService.On("GenerateRandomBytes").Return(tokenBytes, nil)
	s.oneTimeTokenRepository.On("Create", mock.Anything, mock.MatchedBy(func(token model.OneTimeTokenModel) bool {
		return token.UserID == expectedOneTimeToken.UserID &&
			token.TokenType == expectedOneTimeToken.TokenType &&
			string(token.TokenHash) == string(expectedOneTimeToken.TokenHash) &&
			token.ExpiresAt.Sub(expectedOneTimeToken.ExpiresAt) < time.Second // Allow for small time differences
	})).Return(createdToken, nil)
	s.sendEmailConfirmationService.On("Execute", mock.Anything, mock.MatchedBy(func(input service.SendEmailConfirmationInput) bool {
		return input.UserModel.ID == user.ID &&
			string(input.ConfirmationTokenHash) == string(tokenBytes)
	})).
		Return(nil)

	// Act
	err = s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().NoError(err)
	s.sendEmailConfirmationService.AssertExpectations(s.T())
	s.oneTimeTokenRepository.AssertExpectations(s.T())
	s.userRepository.AssertExpectations(s.T())
	s.hashService.AssertExpectations(s.T())
}

func (s *UserCreatedConsumerTestSuite) TestProcessMessage_InvalidJSON_ReturnsError() {
	// Arrange
	ctx := context.Background()
	message := kafka.Message{
		Value: []byte("invalid-json"),
	}

	// Act
	err := s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().Error(err)
}

func (s *UserCreatedConsumerTestSuite) TestProcessMessage_ZeroUserID_ReturnsError() {
	// Arrange
	ctx := context.Background()

	userCreatedMessage := event.UserCreatedMessage{
		UserID: 0, // Invalid user ID
	}
	messageBytes, err := json.Marshal(userCreatedMessage)
	s.Require().NoError(err)

	message := kafka.Message{
		Value: messageBytes,
	}

	// Act
	err = s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid user ID")
}

func (s *UserCreatedConsumerTestSuite) TestProcessMessage_UserNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userCreatedMessage := event.UserCreatedMessage{
		UserID: userID,
	}
	messageBytes, err := json.Marshal(userCreatedMessage)
	s.Require().NoError(err)

	message := kafka.Message{
		Value: messageBytes,
	}

	userNotFoundError := errors.New("user not found")

	// Mock expectations
	s.userRepository.On("FindByID", mock.Anything, userID).Return(model.UserModel{}, userNotFoundError)

	// Act
	err = s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().Error(err)
	s.Equal(userNotFoundError, err)
	s.userRepository.AssertExpectations(s.T())
}

func (s *UserCreatedConsumerTestSuite) TestProcessMessage_HashServiceFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userCreatedMessage := event.UserCreatedMessage{
		UserID: userID,
	}
	messageBytes, err := json.Marshal(userCreatedMessage)
	s.Require().NoError(err)

	message := kafka.Message{
		Value: messageBytes,
	}

	user := model.UserModel{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	hashError := errors.New("failed to generate random bytes")

	// Mock expectations
	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.hashService.On("GenerateRandomBytes").Return(nil, hashError)

	// Act
	err = s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().Error(err)
	s.Equal(hashError, err)
	s.userRepository.AssertExpectations(s.T())
	s.hashService.AssertExpectations(s.T())
}

func (s *UserCreatedConsumerTestSuite) TestProcessMessage_OneTimeTokenCreationFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userCreatedMessage := event.UserCreatedMessage{
		UserID: userID,
	}
	messageBytes, err := json.Marshal(userCreatedMessage)
	s.Require().NoError(err)

	message := kafka.Message{
		Value: messageBytes,
	}

	user := model.UserModel{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	tokenBytes := []byte("random-token-bytes")
	tokenCreationError := errors.New("failed to create one-time token")

	// Mock expectations
	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.hashService.On("GenerateRandomBytes").Return(tokenBytes, nil)
	s.oneTimeTokenRepository.On("Create", mock.Anything, mock.AnythingOfType("model.OneTimeTokenModel")).
		Return(model.OneTimeTokenModel{}, tokenCreationError)

	// Act
	err = s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().Error(err)
	s.Equal(tokenCreationError, err)
	s.userRepository.AssertExpectations(s.T())
	s.hashService.AssertExpectations(s.T())
	s.oneTimeTokenRepository.AssertExpectations(s.T())
}

func (s *UserCreatedConsumerTestSuite) TestProcessMessage_SendEmailConfirmationFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userCreatedMessage := event.UserCreatedMessage{
		UserID: userID,
	}
	messageBytes, err := json.Marshal(userCreatedMessage)
	s.Require().NoError(err)

	message := kafka.Message{
		Value: messageBytes,
	}

	user := model.UserModel{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	tokenBytes := []byte("random-token-bytes")
	createdToken := model.OneTimeTokenModel{
		ID:        1,
		UserID:    user.ID,
		TokenType: enum.TokenTypeAccountConfirmation,
		TokenHash: tokenBytes,
		CreatedAt: time.Now(),
	}

	emailError := errors.New("failed to send confirmation email")

	// Mock expectations
	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.hashService.On("GenerateRandomBytes").Return(tokenBytes, nil)
	s.oneTimeTokenRepository.On("Create", mock.Anything, mock.AnythingOfType("model.OneTimeTokenModel")).
		Return(createdToken, nil)
	s.sendEmailConfirmationService.On("Execute", mock.Anything, mock.Anything).Return(emailError)

	// Act
	err = s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().Error(err)
	s.Equal(emailError, err)
	s.userRepository.AssertExpectations(s.T())
	s.hashService.AssertExpectations(s.T())
	s.oneTimeTokenRepository.AssertExpectations(s.T())
	s.sendEmailConfirmationService.AssertExpectations(s.T())
}
