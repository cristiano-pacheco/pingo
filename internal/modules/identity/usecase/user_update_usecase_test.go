package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/mocks"
	service_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/service/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/usecase"
	validator_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/validator/mocks"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	shared_validator_mocks "github.com/cristiano-pacheco/pingo/internal/shared/modules/validator/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UserUpdateUseCaseTestSuite struct {
	suite.Suite
	sut                   *usecase.UserUpdateUseCase
	hashServiceMock       *service_mocks.MockHashServiceI
	userRepositoryMock    *mocks.MockUserRepositoryI
	validatorMock         *shared_validator_mocks.MockValidate
	passwordValidatorMock *validator_mocks.MockPasswordValidatorI
	logger                logger.Logger
	cfg                   config.Config
}

func (s *UserUpdateUseCaseTestSuite) SetupTest() {
	s.hashServiceMock = service_mocks.NewMockHashServiceI(s.T())
	s.userRepositoryMock = mocks.NewMockUserRepositoryI(s.T())
	s.validatorMock = shared_validator_mocks.NewMockValidate(s.T())
	s.passwordValidatorMock = validator_mocks.NewMockPasswordValidatorI(s.T())

	s.cfg = config.Config{
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

	s.logger = logger.New(s.cfg)

	s.sut = usecase.NewUserUpdateUseCase(
		s.hashServiceMock,
		s.userRepositoryMock,
		s.validatorMock,
		s.passwordValidatorMock,
		s.logger,
	)
}

func TestUserUpdateUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UserUpdateUseCaseTestSuite))
}

func (s *UserUpdateUseCaseTestSuite) TestExecute_ValidInputWithoutPassword_UpdatesUserSuccessfully() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserUpdateInput{
		UserID:    123,
		FirstName: "John",
		LastName:  "Doe",
	}

	existingUser := model.UserModel{
		ID:        123,
		FirstName: "OldFirst",
		LastName:  "OldLast",
		Email:     "john.doe@example.com",
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, uint64(123)).Return(existingUser, nil)
	s.userRepositoryMock.On("Update", mock.Anything, mock.MatchedBy(func(user model.UserModel) bool {
		return user.ID == 123 && user.FirstName == "John" && user.LastName == "Doe"
	})).Return(nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
}

func (s *UserUpdateUseCaseTestSuite) TestExecute_ValidInputWithPassword_UpdatesUserSuccessfully() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserUpdateInput{
		UserID:    123,
		FirstName: "John",
		LastName:  "Doe",
		Password:  "NewPassword123!",
	}

	existingUser := model.UserModel{
		ID:        123,
		FirstName: "OldFirst",
		LastName:  "OldLast",
		Email:     "john.doe@example.com",
	}

	passwordHash := []byte("hashed-password")

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, uint64(123)).Return(existingUser, nil)
	s.passwordValidatorMock.On("Validate", "NewPassword123!").Return(nil)
	s.hashServiceMock.On("GenerateFromPassword", []byte("NewPassword123!")).Return(passwordHash, nil)
	s.userRepositoryMock.On("Update", mock.Anything, mock.MatchedBy(func(user model.UserModel) bool {
		return user.ID == 123 &&
			user.FirstName == "John" &&
			user.LastName == "Doe" &&
			string(user.PasswordHash) == string(passwordHash)
	})).Return(nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
}

func (s *UserUpdateUseCaseTestSuite) TestExecute_ValidationFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserUpdateInput{
		UserID:    123,
		FirstName: "",
		LastName:  "",
	}
	validationError := errors.New("validation error")

	s.validatorMock.On("Struct", input).Return(validationError)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(validationError, err)
}

func (s *UserUpdateUseCaseTestSuite) TestExecute_UserNotFound_ReturnsUserNotFoundError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserUpdateInput{
		UserID:    999,
		FirstName: "John",
		LastName:  "Doe",
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, uint64(999)).
		Return(model.UserModel{}, shared_errs.ErrRecordNotFound)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Require().ErrorIs(err, errs.ErrUserNotFound)
}

func (s *UserUpdateUseCaseTestSuite) TestExecute_FindByIDRepositoryError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserUpdateInput{
		UserID:    123,
		FirstName: "John",
		LastName:  "Doe",
	}
	repositoryError := errors.New("database error")

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, uint64(123)).Return(model.UserModel{}, repositoryError)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(repositoryError, err)
}

func (s *UserUpdateUseCaseTestSuite) TestExecute_PasswordValidationFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserUpdateInput{
		UserID:    123,
		FirstName: "John",
		LastName:  "Doe",
		Password:  "weak",
	}

	existingUser := model.UserModel{
		ID:        123,
		FirstName: "OldFirst",
		LastName:  "OldLast",
		Email:     "john.doe@example.com",
	}

	passwordValidationError := errors.New("password too weak")

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, uint64(123)).Return(existingUser, nil)
	s.passwordValidatorMock.On("Validate", "weak").Return(passwordValidationError)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(passwordValidationError, err)
}

func (s *UserUpdateUseCaseTestSuite) TestExecute_GenerateFromPasswordFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserUpdateInput{
		UserID:    123,
		FirstName: "John",
		LastName:  "Doe",
		Password:  "ValidPassword123!",
	}

	existingUser := model.UserModel{
		ID:        123,
		FirstName: "OldFirst",
		LastName:  "OldLast",
		Email:     "john.doe@example.com",
	}

	hashError := errors.New("hash generation error")

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, uint64(123)).Return(existingUser, nil)
	s.passwordValidatorMock.On("Validate", "ValidPassword123!").Return(nil)
	s.hashServiceMock.On("GenerateFromPassword", []byte("ValidPassword123!")).Return([]byte{}, hashError)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(hashError, err)
}

func (s *UserUpdateUseCaseTestSuite) TestExecute_UpdateRepositoryFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserUpdateInput{
		UserID:    123,
		FirstName: "John",
		LastName:  "Doe",
	}

	existingUser := model.UserModel{
		ID:        123,
		FirstName: "OldFirst",
		LastName:  "OldLast",
		Email:     "john.doe@example.com",
	}

	updateError := errors.New("update error")

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, uint64(123)).Return(existingUser, nil)
	s.userRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.UserModel")).Return(updateError)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(updateError, err)
}
