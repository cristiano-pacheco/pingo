package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	"github.com/stretchr/testify/suite"
)

type EmailTemplateServiceTestSuite struct {
	suite.Suite
	sut              service.EmailTemplateService
	originalWorkDir  string
	projectRootFound bool
}

func (s *EmailTemplateServiceTestSuite) SetupSuite() {
	// Get current working directory
	currentDir, err := os.Getwd()
	s.Require().NoError(err)
	s.originalWorkDir = currentDir

	// Find project root by looking for go.mod
	s.projectRootFound = s.findAndSetProjectRoot(currentDir)
}

func (s *EmailTemplateServiceTestSuite) TearDownSuite() {
	// Restore original working directory
	if s.originalWorkDir != "" {
		os.Chdir(s.originalWorkDir)
	}
}

func (s *EmailTemplateServiceTestSuite) findAndSetProjectRoot(startPath string) bool {
	currentPath := startPath

	for {
		// Check if go.mod exists in current directory
		goModPath := filepath.Join(currentPath, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod, set this as working directory
			err = os.Chdir(currentPath)
			if err == nil {
				return true
			}
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached root directory, stop searching
			break
		}
		currentPath = parentPath
	}

	return false
}

func (s *EmailTemplateServiceTestSuite) SetupTest() {
	s.sut = service.NewEmailTemplateService()
}

func TestEmailTemplateServiceSuite(t *testing.T) {
	suite.Run(t, new(EmailTemplateServiceTestSuite))
}

func (s *EmailTemplateServiceTestSuite) TestCompileAccountConfirmationTemplate_ValidInput_ReturnsCompiledHTML() {
	// Skip test if project root not found
	if !s.projectRootFound {
		s.T().Skip("Project root not found, skipping template tests")
	}

	// Arrange
	input := service.AccountConfirmationInput{
		Name:                    "John Doe",
		AccountConfirmationLink: "https://example.com/confirm?token=abc123",
	}

	// Act
	result, err := s.sut.CompileAccountConfirmationTemplate(input)

	// Assert
	s.NoError(err)
	s.NotEmpty(result)
	s.Contains(result, "John Doe")
	s.Contains(result, "https://example.com/confirm?token=abc123")
	s.Contains(result, "Account Confirmation")
	s.Contains(result, "<!DOCTYPE html>")
	s.Contains(result, "Confirm my account")
}

func (s *EmailTemplateServiceTestSuite) TestCompileAccountConfirmationTemplate_EmptyName_ReturnsCompiledHTML() {
	// Skip test if project root not found
	if !s.projectRootFound {
		s.T().Skip("Project root not found, skipping template tests")
	}

	// Arrange
	input := service.AccountConfirmationInput{
		Name:                    "",
		AccountConfirmationLink: "https://example.com/confirm?token=abc123",
	}

	// Act
	result, err := s.sut.CompileAccountConfirmationTemplate(input)

	// Assert
	s.NoError(err)
	s.NotEmpty(result)
	s.Contains(result, "https://example.com/confirm?token=abc123")
	s.Contains(result, "Account Confirmation")
	s.Contains(result, "<!DOCTYPE html>")
}

func (s *EmailTemplateServiceTestSuite) TestCompileAccountConfirmationTemplate_EmptyLink_ReturnsCompiledHTML() {
	// Skip test if project root not found
	if !s.projectRootFound {
		s.T().Skip("Project root not found, skipping template tests")
	}

	// Arrange
	input := service.AccountConfirmationInput{
		Name:                    "John Doe",
		AccountConfirmationLink: "",
	}

	// Act
	result, err := s.sut.CompileAccountConfirmationTemplate(input)

	// Assert
	s.NoError(err)
	s.NotEmpty(result)
	s.Contains(result, "John Doe")
	s.Contains(result, "Account Confirmation")
	s.Contains(result, "<!DOCTYPE html>")
}

func (s *EmailTemplateServiceTestSuite) TestCompileAccountConfirmationTemplate_InvalidTemplatePath_ReturnsError() {
	// Arrange
	originalWd, _ := os.Getwd()
	tempDir := s.T().TempDir()
	err := os.Chdir(tempDir)
	s.Require().NoError(err)

	defer func() {
		os.Chdir(originalWd)
	}()

	input := service.AccountConfirmationInput{
		Name:                    "John Doe",
		AccountConfirmationLink: "https://example.com/confirm",
	}

	// Act
	result, err := s.sut.CompileAccountConfirmationTemplate(input)

	// Assert
	s.Error(err)
	s.Empty(result)
}

func (s *EmailTemplateServiceTestSuite) TestCompileAuthVerificationCodeTemplate_ValidInput_ReturnsCompiledHTML() {
	// Skip test if project root not found
	if !s.projectRootFound {
		s.T().Skip("Project root not found, skipping template tests")
	}

	// Arrange
	name := "Jane Smith"
	code := "123456"

	// Act
	result, err := s.sut.CompileAuthVerificationCodeTemplate(name, code)

	// Assert
	s.NoError(err)
	s.NotEmpty(result)
	s.Contains(result, "Jane Smith")
	s.Contains(result, "123456")
	s.Contains(result, "Auth Verification Code")
	s.Contains(result, "<!DOCTYPE html>")
	s.Contains(result, "Your verification code is:")
}

func (s *EmailTemplateServiceTestSuite) TestCompileAuthVerificationCodeTemplate_EmptyName_ReturnsCompiledHTML() {
	// Skip test if project root not found
	if !s.projectRootFound {
		s.T().Skip("Project root not found, skipping template tests")
	}

	// Arrange
	name := ""
	code := "123456"

	// Act
	result, err := s.sut.CompileAuthVerificationCodeTemplate(name, code)

	// Assert
	s.NoError(err)
	s.NotEmpty(result)
	s.Contains(result, "123456")
	s.Contains(result, "Auth Verification Code")
	s.Contains(result, "<!DOCTYPE html>")
}

func (s *EmailTemplateServiceTestSuite) TestCompileAuthVerificationCodeTemplate_EmptyCode_ReturnsCompiledHTML() {
	// Skip test if project root not found
	if !s.projectRootFound {
		s.T().Skip("Project root not found, skipping template tests")
	}

	// Arrange
	name := "Jane Smith"
	code := ""

	// Act
	result, err := s.sut.CompileAuthVerificationCodeTemplate(name, code)

	// Assert
	s.NoError(err)
	s.NotEmpty(result)
	s.Contains(result, "Jane Smith")
	s.Contains(result, "Auth Verification Code")
	s.Contains(result, "<!DOCTYPE html>")
}

func (s *EmailTemplateServiceTestSuite) TestCompileAuthVerificationCodeTemplate_SpecialCharactersInInputs_ReturnsCompiledHTML() {
	// Skip test if project root not found
	if !s.projectRootFound {
		s.T().Skip("Project root not found, skipping template tests")
	}

	// Arrange
	name := "José García & María"
	code := "A1B2C3"

	// Act
	result, err := s.sut.CompileAuthVerificationCodeTemplate(name, code)

	// Assert
	s.NoError(err)
	s.NotEmpty(result)
	// HTML templates escape special characters automatically
	s.True(strings.Contains(result, "José García") && (strings.Contains(result, "&amp;") || strings.Contains(result, "&")))
	s.Contains(result, "A1B2C3")
	s.Contains(result, "Auth Verification Code")
	s.Contains(result, "<!DOCTYPE html>")
}

func (s *EmailTemplateServiceTestSuite) TestCompileAuthVerificationCodeTemplate_InvalidTemplatePath_ReturnsError() {
	// Arrange
	originalWd, _ := os.Getwd()
	tempDir := s.T().TempDir()
	err := os.Chdir(tempDir)
	s.Require().NoError(err)

	defer func() {
		os.Chdir(originalWd)
	}()

	name := "Jane Smith"
	code := "123456"

	// Act
	result, err := s.sut.CompileAuthVerificationCodeTemplate(name, code)

	// Assert
	s.Error(err)
	s.Empty(result)
}

func (s *EmailTemplateServiceTestSuite) TestCompileAccountConfirmationTemplate_LongInputValues_ReturnsCompiledHTML() {
	// Skip test if project root not found
	if !s.projectRootFound {
		s.T().Skip("Project root not found, skipping template tests")
	}

	// Arrange
	longName := "This is a very long name that contains multiple words and special characters like áéíóú çñ"
	longLink := "https://example.com/confirm?token=" +
		"verylongtokenvaluewithhexadecimalcharacterslike1234567890abcdef" +
		"&redirect=https://app.example.com/dashboard"

	input := service.AccountConfirmationInput{
		Name:                    longName,
		AccountConfirmationLink: longLink,
	}

	// Act
	result, err := s.sut.CompileAccountConfirmationTemplate(input)

	// Assert
	s.NoError(err)
	s.NotEmpty(result)
	s.Contains(result, "This is a very long name")
	s.Contains(result, "verylongtokenvalue")
	s.Contains(result, "Account Confirmation")
	s.Contains(result, "<!DOCTYPE html>")
}

func (s *EmailTemplateServiceTestSuite) TestCompileAuthVerificationCodeTemplate_NumericAndAlphanumericCodes_ReturnsCompiledHTML() {
	// Skip test if project root not found
	if !s.projectRootFound {
		s.T().Skip("Project root not found, skipping template tests")
	}

	// Arrange
	testCases := []struct {
		name string
		code string
	}{
		{"User1", "123456"},
		{"User2", "ABC123"},
		{"User3", "a1b2c3"},
		{"User4", "000000"},
		{"User5", "ZZZZZ9"},
	}

	for _, tc := range testCases {
		s.Run("Code_"+tc.code, func() {
			// Act
			result, err := s.sut.CompileAuthVerificationCodeTemplate(tc.name, tc.code)

			// Assert
			s.NoError(err)
			s.NotEmpty(result)
			s.Contains(result, tc.name)
			s.Contains(result, tc.code)
			s.Contains(result, "Auth Verification Code")
			s.Contains(result, "<!DOCTYPE html>")
		})
	}
}
