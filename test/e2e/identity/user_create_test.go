//go:build e2e

package identity_test

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/cristiano-pacheco/pingo/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCreateSuccess(t *testing.T) {
	// Arrange
	url := "/api/v1/users"

	// Generate unique email for this test run
	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("test%d@gmail.com", timestamp)

	// Create request body
	requestBody := map[string]interface{}{
		"first_name": "cristiano",
		"last_name":  "pacheco",
		"email":      email,
		"password":   "Ci@23456789",
	}

	headers := map[string]string{
		"Accept": "application/json",
	}

	// Act
	resp, err := test.MakeRequest("POST", url, requestBody, headers)
	require.NoError(t, err)

	defer resp.Body.Close()

	resbody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 201, resp.StatusCode)

	// Parse and check response structure
	var response struct {
		Data struct {
			UserID    int    `json:"user_id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
		} `json:"data"`
	}

	err = json.Unmarshal(resbody, &response)
	require.NoError(t, err)

	assert.Equal(t, "cristiano", response.Data.FirstName)
	assert.Equal(t, "pacheco", response.Data.LastName)
	assert.Equal(t, email, response.Data.Email)
	assert.Greater(t, response.Data.UserID, 0)
}

func TestUserCreateValidation(t *testing.T) {
	url := "/api/v1/users"

	testCases := []struct {
		name               string
		requestBody        map[string]interface{}
		expectedStatusCode int
		expectedMessage    string
	}{
		// FirstName validation tests
		{
			name: "FirstName_Required",
			requestBody: map[string]interface{}{
				"last_name": "pacheco",
				"email":     "test@gmail.com",
				"password":  "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "FirstName_Empty",
			requestBody: map[string]interface{}{
				"first_name": "",
				"last_name":  "pacheco",
				"email":      "test@gmail.com",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "FirstName_TooShort",
			requestBody: map[string]interface{}{
				"first_name": "ab",
				"last_name":  "pacheco",
				"email":      "test@gmail.com",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "FirstName_TooLong",
			requestBody: map[string]interface{}{
				"first_name": strings.Repeat("a", 256),
				"last_name":  "pacheco",
				"email":      "test@gmail.com",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		// LastName validation tests
		{
			name: "LastName_Required",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"email":      "test@gmail.com",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "LastName_Empty",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  "",
				"email":      "test@gmail.com",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "LastName_TooShort",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  "ab",
				"email":      "test@gmail.com",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "LastName_TooLong",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  strings.Repeat("a", 256),
				"email":      "test@gmail.com",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		// Password validation tests
		{
			name: "Password_Required",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  "pacheco",
				"email":      "test@gmail.com",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Password_Empty",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  "pacheco",
				"email":      "test@gmail.com",
				"password":   "",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Password_TooShort",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  "pacheco",
				"email":      "test@gmail.com",
				"password":   "1234567",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		// Email validation tests
		{
			name: "Email_Required",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  "pacheco",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Email_Empty",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  "pacheco",
				"email":      "",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Email_InvalidFormat_NoAtSymbol",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  "pacheco",
				"email":      "testgmail.com",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Email_InvalidFormat_NoDomain",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  "pacheco",
				"email":      "test@",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Email_InvalidFormat_NoUsername",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  "pacheco",
				"email":      "@gmail.com",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Email_TooLong",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  "pacheco",
				"email":      strings.Repeat("a", 250) + "@gmail.com", // 250 + "@gmail.com" = 260 characters
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
	}

	headers := map[string]string{
		"Accept": "application/json",
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate unique email for tests that need valid email format
			if email, ok := tc.requestBody["email"].(string); ok && email == "test@gmail.com" {
				timestamp := time.Now().UnixNano()
				tc.requestBody["email"] = fmt.Sprintf("test%d@gmail.com", timestamp)
			}

			// Act
			resp, err := test.MakeRequest("POST", url, tc.requestBody, headers)
			require.NoError(t, err)

			defer resp.Body.Close()

			resbody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode, "Response body: %s", string(resbody))

			if tc.expectedStatusCode == 400 {
				// Parse error response to check message
				var errorResponse struct {
					Message string `json:"message"`
				}

				err = json.Unmarshal(resbody, &errorResponse)
				require.NoError(t, err)

				assert.Contains(t, errorResponse.Message, tc.expectedMessage)
			}
		})
	}
}

func TestUserCreateValidation_EdgeCases(t *testing.T) {
	url := "/api/v1/users"

	testCases := []struct {
		name               string
		requestBody        map[string]interface{}
		expectedStatusCode int
		expectedMessage    string
	}{
		// Edge cases for minimum valid values
		{
			name: "FirstName_MinimumValid",
			requestBody: map[string]interface{}{
				"first_name": "abc",
				"last_name":  "def",
				"email":      "test@gmail.com",
				"password":   "12345678",
			},
			expectedStatusCode: 201, // Should succeed
		},
		{
			name: "LastName_MinimumValid",
			requestBody: map[string]interface{}{
				"first_name": "abc",
				"last_name":  "def",
				"email":      "test@gmail.com",
				"password":   "12345678",
			},
			expectedStatusCode: 201, // Should succeed
		},
		{
			name: "Password_MinimumValid",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  "pacheco",
				"email":      "test@gmail.com",
				"password":   "12345678",
			},
			expectedStatusCode: 201, // Should succeed
		},
		// Edge cases for maximum valid values
		{
			name: "FirstName_MaximumValid",
			requestBody: map[string]interface{}{
				"first_name": strings.Repeat("a", 255),
				"last_name":  "pacheco",
				"email":      "test@gmail.com",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 201, // Should succeed
		},
		{
			name: "LastName_MaximumValid",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  strings.Repeat("a", 255),
				"email":      "test@gmail.com",
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 201, // Should succeed
		},
		{
			name: "Email_MaximumValid",
			requestBody: map[string]interface{}{
				"first_name": "cristiano",
				"last_name":  "pacheco",
				"email":      strings.Repeat("a", 246) + "@test.com", // 246 + "@test.com" = 255 characters
				"password":   "Ci@23456789",
			},
			expectedStatusCode: 201, // Should succeed
		},
	}

	headers := map[string]string{
		"Accept": "application/json",
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate unique email for success cases
			if tc.expectedStatusCode == 201 {
				timestamp := time.Now().UnixNano()

				// For the maximum email test, create a unique long email
				if strings.Contains(tc.name, "Email_MaximumValid") {
					uniqueLongEmail := fmt.Sprintf("%s%d@test.com", strings.Repeat("a", 240), timestamp)
					tc.requestBody["email"] = uniqueLongEmail
				} else {
					tc.requestBody["email"] = fmt.Sprintf("test%d@gmail.com", timestamp)
				}
			}

			// Act
			resp, err := test.MakeRequest("POST", url, tc.requestBody, headers)
			require.NoError(t, err)

			defer resp.Body.Close()

			resbody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode, "Response body: %s", string(resbody))

			if tc.expectedStatusCode == 201 {
				// Parse success response
				var response struct {
					Data struct {
						UserID    int    `json:"user_id"`
						FirstName string `json:"first_name"`
						LastName  string `json:"last_name"`
						Email     string `json:"email"`
					} `json:"data"`
				}

				err = json.Unmarshal(resbody, &response)
				require.NoError(t, err)

				assert.Greater(t, response.Data.UserID, 0)
			} else if tc.expectedStatusCode == 400 {
				// Parse error response
				var errorResponse struct {
					Message string `json:"message"`
				}

				err = json.Unmarshal(resbody, &errorResponse)
				require.NoError(t, err)

				assert.Contains(t, errorResponse.Message, tc.expectedMessage)
			}
		})
	}
}
