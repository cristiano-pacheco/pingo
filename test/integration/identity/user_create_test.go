//go:build integration

package identity_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cristiano-pacheco/pingo/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCreateSuccess(t *testing.T) {
	// Arrange
	apiURL := test.GetAPIBaseUrl()
	url := apiURL + "/api/v1/users"

	// Generate unique email for this test run
	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("test%d@gmail.com", timestamp)

	// Create request body
	requestBody := fmt.Sprintf(`{
		"first_name": "cristiano",
		"last_name": "pacheco",
		"email": "%s",
		"password": "Ci@23456789"
	}`, email)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(requestBody))
	require.NoError(t, err)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	// Act
	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	resbody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Assert
	require.Equal(t, http.StatusCreated, resp.StatusCode)

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
	apiURL := test.GetAPIBaseUrl()
	url := apiURL + "/api/v1/users"

	testCases := []struct {
		name               string
		requestBody        string
		expectedStatusCode int
		expectedMessage    string
	}{
		// FirstName validation tests
		{
			name: "FirstName_Required",
			requestBody: `{
				"last_name": "pacheco",
				"email": "test@gmail.com",
				"password": "Ci@23456789"
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "FirstName_Empty",
			requestBody: `{
				"first_name": "",
				"last_name": "pacheco",
				"email": "test@gmail.com",
				"password": "Ci@23456789"
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "FirstName_TooShort",
			requestBody: `{
				"first_name": "ab",
				"last_name": "pacheco",
				"email": "test@gmail.com",
				"password": "Ci@23456789"
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "FirstName_TooLong",
			requestBody: fmt.Sprintf(`{
				"first_name": "%s",
				"last_name": "pacheco",
				"email": "test@gmail.com",
				"password": "Ci@23456789"
			}`, strings.Repeat("a", 256)),
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		// LastName validation tests
		{
			name: "LastName_Required",
			requestBody: `{
				"first_name": "cristiano",
				"email": "test@gmail.com",
				"password": "Ci@23456789"
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "LastName_Empty",
			requestBody: `{
				"first_name": "cristiano",
				"last_name": "",
				"email": "test@gmail.com",
				"password": "Ci@23456789"
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "LastName_TooShort",
			requestBody: `{
				"first_name": "cristiano",
				"last_name": "ab",
				"email": "test@gmail.com",
				"password": "Ci@23456789"
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "LastName_TooLong",
			requestBody: fmt.Sprintf(`{
				"first_name": "cristiano",
				"last_name": "%s",
				"email": "test@gmail.com",
				"password": "Ci@23456789"
			}`, strings.Repeat("a", 256)),
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		// Password validation tests
		{
			name: "Password_Required",
			requestBody: `{
				"first_name": "cristiano",
				"last_name": "pacheco",
				"email": "test@gmail.com"
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Password_Empty",
			requestBody: `{
				"first_name": "cristiano",
				"last_name": "pacheco",
				"email": "test@gmail.com",
				"password": ""
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Password_TooShort",
			requestBody: `{
				"first_name": "cristiano",
				"last_name": "pacheco",
				"email": "test@gmail.com",
				"password": "1234567"
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		// Email validation tests
		{
			name: "Email_Required",
			requestBody: `{
				"first_name": "cristiano",
				"last_name": "pacheco",
				"password": "Ci@23456789"
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Email_Empty",
			requestBody: `{
				"first_name": "cristiano",
				"last_name": "pacheco",
				"email": "",
				"password": "Ci@23456789"
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Email_InvalidFormat_NoAtSymbol",
			requestBody: `{
				"first_name": "cristiano",
				"last_name": "pacheco",
				"email": "testgmail.com",
				"password": "Ci@23456789"
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Email_InvalidFormat_NoDomain",
			requestBody: `{
				"first_name": "cristiano",
				"last_name": "pacheco",
				"email": "test@",
				"password": "Ci@23456789"
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Email_InvalidFormat_NoUsername",
			requestBody: `{
				"first_name": "cristiano",
				"last_name": "pacheco",
				"email": "@gmail.com",
				"password": "Ci@23456789"
			}`,
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
		{
			name: "Email_TooLong",
			requestBody: fmt.Sprintf(`{
				"first_name": "cristiano",
				"last_name": "pacheco",
				"email": "%s@gmail.com",
				"password": "Ci@23456789"
			}`, strings.Repeat("a", 250)), // 250 + "@gmail.com" = 260 characters
			expectedStatusCode: 400,
			expectedMessage:    "validation failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate unique email for tests that need valid email format
			if strings.Contains(tc.requestBody, "test@gmail.com") {
				timestamp := time.Now().UnixNano()
				uniqueEmail := fmt.Sprintf("test%d@gmail.com", timestamp)
				tc.requestBody = strings.ReplaceAll(tc.requestBody, "test@gmail.com", uniqueEmail)
			}

			req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(tc.requestBody))
			require.NoError(t, err)

			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}

			// Act
			resp, err := client.Do(req)
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
	apiURL := test.GetAPIBaseUrl()
	url := apiURL + "/api/v1/users"

	testCases := []struct {
		name               string
		requestBody        string
		expectedStatusCode int
		expectedMessage    string
	}{
		// Edge cases for minimum valid values
		{
			name: "FirstName_MinimumValid",
			requestBody: `{
				"first_name": "abc",
				"last_name": "def",
				"email": "test@gmail.com",
				"password": "12345678"
			}`,
			expectedStatusCode: 201, // Should succeed
		},
		{
			name: "LastName_MinimumValid",
			requestBody: `{
				"first_name": "abc",
				"last_name": "def",
				"email": "test@gmail.com",
				"password": "12345678"
			}`,
			expectedStatusCode: 201, // Should succeed
		},
		{
			name: "Password_MinimumValid",
			requestBody: `{
				"first_name": "cristiano",
				"last_name": "pacheco",
				"email": "test@gmail.com",
				"password": "12345678"
			}`,
			expectedStatusCode: 201, // Should succeed
		},
		// Edge cases for maximum valid values
		{
			name: "FirstName_MaximumValid",
			requestBody: fmt.Sprintf(`{
				"first_name": "%s",
				"last_name": "pacheco",
				"email": "test@gmail.com",
				"password": "Ci@23456789"
			}`, strings.Repeat("a", 255)),
			expectedStatusCode: 201, // Should succeed
		},
		{
			name: "LastName_MaximumValid",
			requestBody: fmt.Sprintf(`{
				"first_name": "cristiano",
				"last_name": "%s",
				"email": "test@gmail.com",
				"password": "Ci@23456789"
			}`, strings.Repeat("a", 255)),
			expectedStatusCode: 201, // Should succeed
		},
		{
			name: "Email_MaximumValid",
			requestBody: fmt.Sprintf(`{
				"first_name": "cristiano",
				"last_name": "pacheco",
				"email": "%s@test.com",
				"password": "Ci@23456789"
			}`, strings.Repeat("a", 246)), // 246 + "@test.com" = 255 characters
			expectedStatusCode: 201, // Should succeed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate unique email for success cases
			if tc.expectedStatusCode == 201 {
				timestamp := time.Now().UnixNano()
				uniqueEmail := fmt.Sprintf("test%d@gmail.com", timestamp)
				tc.requestBody = strings.ReplaceAll(tc.requestBody, "test@gmail.com", uniqueEmail)

				// For the maximum email test, create a unique long email
				if strings.Contains(tc.name, "Email_MaximumValid") {
					uniqueLongEmail := fmt.Sprintf("%s%d@test.com", strings.Repeat("a", 240), timestamp)
					tc.requestBody = strings.ReplaceAll(tc.requestBody, strings.Repeat("a", 246)+"@test.com", uniqueLongEmail)
				}
			}

			req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(tc.requestBody))
			require.NoError(t, err)

			req.Header.Set("Accept", "application/json")
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}

			// Act
			resp, err := client.Do(req)
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
