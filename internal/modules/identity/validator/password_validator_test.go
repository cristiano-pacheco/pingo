package validator_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/validator"
)

type PasswordValidatorSuite struct {
	suite.Suite
	sut validator.PasswordValidator
}

func (s *PasswordValidatorSuite) SetupTest() {
	s.sut = validator.NewPasswordValidator()
}

func TestPasswordValidatorSuite(t *testing.T) {
	suite.Run(t, new(PasswordValidatorSuite))
}

func (s *PasswordValidatorSuite) TestValidate_ValidPassword() {
	// Arrange
	password := "Abcdef1!"
	// Act
	err := s.sut.Validate(password)
	// Assert
	s.Require().NoError(err)
}

func (s *PasswordValidatorSuite) TestValidate_PasswordTooShort_ReturnsError() {
	// Arrange
	password := "Ab1!"
	// Act
	err := s.sut.Validate(password)
	// Assert
	s.Require().ErrorIs(err, errs.ErrPasswordTooShort)
}

func (s *PasswordValidatorSuite) TestValidate_NoUppercase_ReturnsError() {
	// Arrange
	password := "abcdef1!"
	// Act
	err := s.sut.Validate(password)
	// Assert
	s.Require().ErrorIs(err, errs.ErrPasswordNoUppercase)
}

func (s *PasswordValidatorSuite) TestValidate_NoLowercase_ReturnsError() {
	// Arrange
	password := "ABCDEF1!"
	// Act
	err := s.sut.Validate(password)
	// Assert
	s.Require().ErrorIs(err, errs.ErrPasswordNoLowercase)
}

func (s *PasswordValidatorSuite) TestValidate_NoNumber_ReturnsError() {
	// Arrange
	password := "Abcdefgh!"
	// Act
	err := s.sut.Validate(password)
	// Assert
	s.Require().ErrorIs(err, errs.ErrPasswordNoNumber)
}

func (s *PasswordValidatorSuite) TestValidate_NoSpecialChar_ReturnsError() {
	// Arrange
	password := "Abcdef12"
	// Act
	err := s.sut.Validate(password)
	// Assert
	s.Require().ErrorIs(err, errs.ErrPasswordNoSpecialChar)
}
