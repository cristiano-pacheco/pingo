package consumer_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event/consumer"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	repository_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	service_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/service/mocks"
	"github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
)

type UserAuthenticatedConsumerTestSuite struct {
	suite.Suite
	sut                              *consumer.UserAuthenticatedConsumer
	sendEmailVerificationCodeService *service_mocks.MockSendEmailVerificationCodeServiceI
	oneTimeTokenRepository           *repository_mocks.MockOneTimeTokenRepositoryI
	userRepository                   *repository_mocks.MockUserRepositoryI
	hashService                      *service_mocks.MockHashServiceI
	logger                           logger.Logger
}

func (s *UserAuthenticatedConsumerTestSuite) SetupTest() {
	s.sendEmailVerificationCodeService = service_mocks.NewMockSendEmailVerificationCodeServiceI(s.T())
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

	s.sut = consumer.NewUserAuthenticatedConsumer(
		s.sendEmailVerificationCodeService,
		s.oneTimeTokenRepository,
		s.userRepository,
		s.hashService,
		s.logger,
	)
}

func TestUserAuthenticatedConsumerSuite(t *testing.T) {
	suite.Run(t, new(UserAuthenticatedConsumerTestSuite))
}

func (s *UserAuthenticatedConsumerTestSuite) TestTopic_ReturnsCorrectTopic() {
	// Arrange
	expectedTopic := event.IdentityUserAuthenticatedTopic

	// Act
	result := s.sut.Topic()

	// Assert
	s.Equal(expectedTopic, result)
}

func (s *UserAuthenticatedConsumerTestSuite) TestGroupID_ReturnsCorrectGroupID() {
	// Arrange
	expectedGroupID := "default"

	// Act
	result := s.sut.GroupID()

	// Assert
	s.Equal(expectedGroupID, result)
}

func (s *UserAuthenticatedConsumerTestSuite) TestProcessMessage_ValidMessage_ProcessesSuccessfully() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userAuthenticatedMessage := event.UserAuthenticatedMessage{
		UserID: userID,
	}
	messageBytes, err := json.Marshal(userAuthenticatedMessage)
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

	tokenHash := []byte("hashed-token")
	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)

	// Mock expectations
	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepository.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(nil)
	s.hashService.On("GenerateFromPassword", mock.AnythingOfType("[]uint8")).Return(tokenHash, nil)
	s.oneTimeTokenRepository.On("Create", mock.Anything, mock.AnythingOfType("model.OneTimeTokenModel")).
		Return(model.OneTimeTokenModel{}, nil)
	s.sendEmailVerificationCodeService.On("Execute", mock.Anything, mock.AnythingOfType("service.SendEmailVerificationCodeInput")).
		Return(nil)

	// Act
	err = s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().NoError(err)
}

func (s *UserAuthenticatedConsumerTestSuite) TestProcessMessage_InvalidMessageFormat_ReturnsError() {
	// Arrange
	ctx := context.Background()
	invalidMessage := kafka.Message{
		Value: []byte("invalid-json"),
	}

	// Act
	err := s.sut.ProcessMessage(ctx, invalidMessage)

	// Assert
	s.Require().Error(err)
}

func (s *UserAuthenticatedConsumerTestSuite) TestProcessMessage_UserNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userAuthenticatedMessage := event.UserAuthenticatedMessage{
		UserID: userID,
	}
	messageBytes, err := json.Marshal(userAuthenticatedMessage)
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
}

func (s *UserAuthenticatedConsumerTestSuite) TestProcessMessage_DeleteTokenRepositoryError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userAuthenticatedMessage := event.UserAuthenticatedMessage{
		UserID: userID,
	}
	messageBytes, err := json.Marshal(userAuthenticatedMessage)
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

	deleteError := errors.New("database error")
	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)

	// Mock expectations
	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepository.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(deleteError)

	// Act
	err = s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().Error(err)
	s.Equal(deleteError, err)
}

func (s *UserAuthenticatedConsumerTestSuite) TestProcessMessage_DeleteTokenNotFound_ContinuesProcessing() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userAuthenticatedMessage := event.UserAuthenticatedMessage{
		UserID: userID,
	}
	messageBytes, err := json.Marshal(userAuthenticatedMessage)
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

	tokenHash := []byte("hashed-token")
	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)

	// Mock expectations
	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepository.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(errs.ErrRecordNotFound)
	s.hashService.On("GenerateFromPassword", mock.AnythingOfType("[]uint8")).Return(tokenHash, nil)
	s.oneTimeTokenRepository.On("Create", mock.Anything, mock.AnythingOfType("model.OneTimeTokenModel")).
		Return(model.OneTimeTokenModel{}, nil)
	s.sendEmailVerificationCodeService.On("Execute", mock.Anything, mock.AnythingOfType("service.SendEmailVerificationCodeInput")).
		Return(nil)

	// Act
	err = s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().NoError(err)
}

func (s *UserAuthenticatedConsumerTestSuite) TestProcessMessage_HashServiceError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userAuthenticatedMessage := event.UserAuthenticatedMessage{
		UserID: userID,
	}
	messageBytes, err := json.Marshal(userAuthenticatedMessage)
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

	hashError := errors.New("hash generation failed")
	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)

	// Mock expectations
	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepository.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(nil)
	s.hashService.On("GenerateFromPassword", mock.AnythingOfType("[]uint8")).Return(nil, hashError)

	// Act
	err = s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().Error(err)
	s.Equal(hashError, err)
}

func (s *UserAuthenticatedConsumerTestSuite) TestProcessMessage_CreateTokenError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userAuthenticatedMessage := event.UserAuthenticatedMessage{
		UserID: userID,
	}
	messageBytes, err := json.Marshal(userAuthenticatedMessage)
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

	tokenHash := []byte("hashed-token")
	createError := errors.New("database error")
	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)

	// Mock expectations
	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepository.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(nil)
	s.hashService.On("GenerateFromPassword", mock.AnythingOfType("[]uint8")).Return(tokenHash, nil)
	s.oneTimeTokenRepository.On("Create", mock.Anything, mock.AnythingOfType("model.OneTimeTokenModel")).
		Return(model.OneTimeTokenModel{}, createError)

	// Act
	err = s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().Error(err)
	s.Equal(createError, err)
}

func (s *UserAuthenticatedConsumerTestSuite) TestProcessMessage_SendEmailError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userAuthenticatedMessage := event.UserAuthenticatedMessage{
		UserID: userID,
	}
	messageBytes, err := json.Marshal(userAuthenticatedMessage)
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

	tokenHash := []byte("hashed-token")
	emailError := errors.New("email sending failed")
	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)

	// Mock expectations
	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepository.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(nil)
	s.hashService.On("GenerateFromPassword", mock.AnythingOfType("[]uint8")).Return(tokenHash, nil)
	s.oneTimeTokenRepository.On("Create", mock.Anything, mock.AnythingOfType("model.OneTimeTokenModel")).
		Return(model.OneTimeTokenModel{}, nil)
	s.sendEmailVerificationCodeService.On("Execute", mock.Anything, mock.AnythingOfType("service.SendEmailVerificationCodeInput")).
		Return(emailError)

	// Act
	err = s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().Error(err)
	s.Equal(emailError, err)
}

func (s *UserAuthenticatedConsumerTestSuite) TestProcessMessage_GeneratesValidVerificationCode_ProcessesSuccessfully() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)

	userAuthenticatedMessage := event.UserAuthenticatedMessage{
		UserID: userID,
	}
	messageBytes, err := json.Marshal(userAuthenticatedMessage)
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

	tokenHash := []byte("hashed-token")
	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)

	// Mock expectations - capture the generated code
	capturedInput := &service.SendEmailVerificationCodeInput{}
	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepository.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(nil)
	s.hashService.On("GenerateFromPassword", mock.AnythingOfType("[]uint8")).Return(tokenHash, nil)
	s.oneTimeTokenRepository.On("Create", mock.Anything, mock.AnythingOfType("model.OneTimeTokenModel")).
		Return(model.OneTimeTokenModel{}, nil)
	s.sendEmailVerificationCodeService.On("Execute", mock.Anything, mock.AnythingOfType("service.SendEmailVerificationCodeInput")).
		Run(func(args mock.Arguments) {
			*capturedInput = args.Get(1).(service.SendEmailVerificationCodeInput)
		}).
		Return(nil)

	// Act
	err = s.sut.ProcessMessage(ctx, message)

	// Assert
	s.Require().NoError(err)
	s.Equal(userID, capturedInput.UserID)
	s.Len(capturedInput.Code, 6)            // Should be a 6-digit code
	s.Regexp(`^\d{6}$`, capturedInput.Code) // Should be exactly 6 digits
}
