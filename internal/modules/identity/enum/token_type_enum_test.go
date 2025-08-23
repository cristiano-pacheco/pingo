package enum_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
)

func TestNewTokenTypeEnum(t *testing.T) {
	t.Run("ValidAccountConfirmationToken_ReturnsValidEnum", func(t *testing.T) {
		// Arrange
		value := enum.TokenTypeAccountConfirmation

		// Act
		result, err := enum.NewTokenTypeEnum(value)

		// Assert
		require.NoError(t, err)
		require.Equal(t, value, result.String())
	})

	t.Run("ValidLoginVerificationToken_ReturnsValidEnum", func(t *testing.T) {
		// Arrange
		value := enum.TokenTypeLoginVerification

		// Act
		result, err := enum.NewTokenTypeEnum(value)

		// Assert
		require.NoError(t, err)
		require.Equal(t, value, result.String())
	})

	t.Run("ValidResetPasswordToken_ReturnsValidEnum", func(t *testing.T) {
		// Arrange
		value := enum.TokenTypeResetPassword

		// Act
		result, err := enum.NewTokenTypeEnum(value)

		// Assert
		require.NoError(t, err)
		require.Equal(t, value, result.String())
	})

	t.Run("InvalidTokenType_ReturnsError", func(t *testing.T) {
		// Arrange
		value := "invalid_token_type"

		// Act
		result, err := enum.NewTokenTypeEnum(value)

		// Assert
		require.ErrorIs(t, err, errs.ErrInvalidTokenType)
		require.Equal(t, "", result.String())
	})

	t.Run("EmptyString_ReturnsError", func(t *testing.T) {
		// Arrange
		value := ""

		// Act
		result, err := enum.NewTokenTypeEnum(value)

		// Assert
		require.ErrorIs(t, err, errs.ErrInvalidTokenType)
		require.Equal(t, "", result.String())
	})

	t.Run("WhitespaceString_ReturnsError", func(t *testing.T) {
		// Arrange
		value := "   "

		// Act
		result, err := enum.NewTokenTypeEnum(value)

		// Assert
		require.ErrorIs(t, err, errs.ErrInvalidTokenType)
		require.Equal(t, "", result.String())
	})
}

func TestTokenTypeEnum_String(t *testing.T) {
	t.Run("AccountConfirmationToken_ReturnsCorrectString", func(t *testing.T) {
		// Arrange
		tokenType, err := enum.NewTokenTypeEnum(enum.TokenTypeAccountConfirmation)
		require.NoError(t, err)

		// Act
		result := tokenType.String()

		// Assert
		require.Equal(t, enum.TokenTypeAccountConfirmation, result)
	})

	t.Run("LoginVerificationToken_ReturnsCorrectString", func(t *testing.T) {
		// Arrange
		tokenType, err := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
		require.NoError(t, err)

		// Act
		result := tokenType.String()

		// Assert
		require.Equal(t, enum.TokenTypeLoginVerification, result)
	})

	t.Run("ResetPasswordToken_ReturnsCorrectString", func(t *testing.T) {
		// Arrange
		tokenType, err := enum.NewTokenTypeEnum(enum.TokenTypeResetPassword)
		require.NoError(t, err)

		// Act
		result := tokenType.String()

		// Assert
		require.Equal(t, enum.TokenTypeResetPassword, result)
	})
}
