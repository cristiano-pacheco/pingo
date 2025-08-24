package enum_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
)

func TestNewUserStatusEnum_ValidStatuses_ReturnsEnum(t *testing.T) {
	t.Run("pending", func(t *testing.T) {
		// Arrange
		val := enum.UserStatusPending
		// Act
		s, err := enum.NewUserStatusEnum(val)
		// Assert
		require.NoError(t, err)
		require.Equal(t, val, s.String())
	})

	t.Run("active", func(t *testing.T) {
		// Arrange
		val := enum.UserStatusActive
		// Act
		s, err := enum.NewUserStatusEnum(val)
		// Assert
		require.NoError(t, err)
		require.Equal(t, val, s.String())
	})

	t.Run("inactive", func(t *testing.T) {
		// Arrange
		val := enum.UserStatusInactive
		// Act
		s, err := enum.NewUserStatusEnum(val)
		// Assert
		require.NoError(t, err)
		require.Equal(t, val, s.String())
	})

	t.Run("suspended", func(t *testing.T) {
		// Arrange
		val := enum.UserStatusSuspended
		// Act
		s, err := enum.NewUserStatusEnum(val)
		// Assert
		require.NoError(t, err)
		require.Equal(t, val, s.String())
	})
}

func TestNewUserStatusEnum_InvalidStatus_ReturnsError(t *testing.T) {
	// Arrange
	invalid := "unknown"
	// Act
	_, err := enum.NewUserStatusEnum(invalid)
	// Assert
	require.ErrorIs(t, err, errs.ErrInvalidUserStatus)
}

func TestUserStatusEnum_String_ReturnsValue(t *testing.T) {
	// Arrange
	val := enum.UserStatusActive
	s, err := enum.NewUserStatusEnum(val)
	require.NoError(t, err)
	// Act
	got := s.String()
	// Assert
	require.Equal(t, val, got)
}
