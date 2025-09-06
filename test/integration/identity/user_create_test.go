//go:build integration

package identity_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCreateSuccess(t *testing.T) {
	// Arrange
	apiURL := os.Getenv("APP_BASE_URL")
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
