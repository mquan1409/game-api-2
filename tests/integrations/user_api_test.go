package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/mquan1409/game-api/internal/config"
	"github.com/mquan1409/game-api/internal/models"
	"github.com/mquan1409/game-api/internal/utils"
)


func TestUserAPI(t *testing.T) {
	baseURL := config.LoadBaseURL()
	// Load test configuration
	cfg := config.LoadConfig("development")

	// Setup
	db, err := utils.SetupTestDB(&cfg)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	// Scan the entire table before tests
	beforeScan, err := utils.ScanEntireTable(db, cfg.TableName)
	if err != nil {
		t.Fatalf("Failed to scan table before tests: %v", err)
	}

	// Test GetUser
	t.Run("GetUser", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/users/user1", baseURL))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var user models.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		assert.NoError(t, err)
		assert.Equal(t, "AliceWonder", user.Username)
		assert.Equal(t, "alice.wonder@example.com", user.Email)
	})

	// Test GetUserBasicsByPrefix
	t.Run("GetUserBasicsByPrefix", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/users?prefix=user", baseURL))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var users []models.UserBasic
		err = json.NewDecoder(resp.Body).Decode(&users)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 3)
		for _, user := range users {
			assert.Contains(t, string(user.UserID), "user")
		}
	})

	// Test GetGameStat
	t.Run("GetGameStat", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/users/user1/games/soccer/stats", baseURL))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var gameStat models.GameStat
		err = json.NewDecoder(resp.Body).Decode(&gameStat)
		assert.NoError(t, err)
		assert.Equal(t, models.UserID("user1"), gameStat.UserID)
		assert.Equal(t, models.GameID("soccer"), gameStat.GameID)
		assert.Equal(t, models.AttributeStat(1), gameStat.GameAttributes["goals"])
	})

	// Test CreateUser
	t.Run("CreateUser", func(t *testing.T) {
		newUser, err := models.NewUser("user6", "TestUser", "test@example.com", []models.GameID{})
		assert.NoError(t, err)
		jsonUser, _ := json.Marshal(newUser)

		resp, err := http.Post(fmt.Sprintf("%s/users", baseURL), "application/json", bytes.NewBuffer(jsonUser))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createdUser models.User
		err = json.NewDecoder(resp.Body).Decode(&createdUser)
		assert.NoError(t, err)
		assert.Equal(t, newUser.Username, createdUser.Username)
		assert.Equal(t, newUser.Email, createdUser.Email)

		// Clean up: Delete the created user
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/users/%s", baseURL, createdUser.UserID), nil)
		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	// Test UpdateUser
	t.Run("UpdateUser", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/users/user2", baseURL))
		assert.NoError(t, err)
		var user models.User
		json.NewDecoder(resp.Body).Decode(&user)

		originalUsername := user.Username
		originalEmail := user.Email

		updatedUser, err := models.NewUser(user.UserID, "UpdatedUser", "updated@example.com", user.GamesPlayed)
		assert.NoError(t, err)
		jsonUser, _ := json.Marshal(updatedUser)

		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/users/user2", baseURL), bytes.NewBuffer(jsonUser))
		req.Header.Set("Content-Type", "application/json")
		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var resultUser models.User
		json.NewDecoder(resp.Body).Decode(&resultUser)
		assert.Equal(t, updatedUser.Username, resultUser.Username)
		assert.Equal(t, updatedUser.Email, resultUser.Email)

		// Clean up: Revert the user to original state
		revertedUser, err := models.NewUser(user.UserID, originalUsername, originalEmail, user.GamesPlayed)
		assert.NoError(t, err)
		jsonUser, _ = json.Marshal(revertedUser)
		req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("%s/users/user2", baseURL), bytes.NewBuffer(jsonUser))
		req.Header.Set("Content-Type", "application/json")
		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test DeleteUser
	t.Run("DeleteUser", func(t *testing.T) {
		newUser, err := models.NewUser("tempuser", "TempUser", "temp@example.com", []models.GameID{})
		assert.NoError(t, err)
		jsonUser, _ := json.Marshal(newUser)

		resp, err := http.Post(fmt.Sprintf("%s/users", baseURL), "application/json", bytes.NewBuffer(jsonUser))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/users/tempuser", baseURL), nil)
		resp, err = http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Verify user is deleted
		resp, err = http.Get(fmt.Sprintf("%s/users/tempuser", baseURL))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	// Scan the entire table after tests
	afterScan, err := utils.ScanEntireTable(db, cfg.TableName)
	if err != nil {
		t.Fatalf("Failed to scan table after tests: %v", err)
	}

	// Compare before and after scans
	if !t.Failed() {
		assert.Equal(t, beforeScan, afterScan, "The database state has changed after running tests")
	}
}
