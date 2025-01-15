package tests

import (
	"net/http"
	"testing"

	"github.com/mquan1409/game-api/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCorsIntegration(t *testing.T) {
	// Load test configuration
	baseURL := config.LoadBaseURL()

	// Test regular GET request
	t.Run("Regular GET Request", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", baseURL+"/games/soccer", nil)
		assert.NoError(t, err)

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "DELETE,GET,OPTIONS,POST,PUT", resp.Header.Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token", resp.Header.Get("Access-Control-Allow-Headers"))
	})

	// Test OPTIONS request
	t.Run("OPTIONS Request", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("OPTIONS", baseURL+"/games/soccer", nil)
		assert.NoError(t, err)

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "DELETE,GET,OPTIONS,POST,PUT", resp.Header.Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token", resp.Header.Get("Access-Control-Allow-Headers"))
	})

	// Test request from different origin
	t.Run("Request with Origin Header", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", baseURL+"/games/soccer", nil)
		assert.NoError(t, err)
		req.Header.Set("Origin", "http://example.com")

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	})
}
