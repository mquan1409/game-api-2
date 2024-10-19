package tests

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/mquan1409/game-api/internal/config"
	"github.com/mquan1409/game-api/internal/utils"
	"github.com/mquan1409/game-api/internal/models"
	"github.com/mquan1409/game-api/internal/services"
	"github.com/mquan1409/game-api/internal/repositories"
)

func TestUserService(t *testing.T) {
	// Load test configuration
	cfg := config.LoadConfig("development")

	// Setup
	db, err := utils.SetupTestDB(&cfg)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	userRepo := repositories.NewDynamoDBUserRepository(db, cfg.TableName)
	gameStatRepo := repositories.NewDynamoDBGameStatRepository(db, cfg.TableName)
	userService := services.NewUserServiceImpl(userRepo, gameStatRepo)

	// Scan the entire table before tests
	beforeScan, err := utils.ScanEntireTable(db, cfg.TableName)
	if err != nil {
		t.Fatalf("Failed to scan table before tests: %v", err)
	}

	// Test GetUser
	t.Run("GetUser", func(t *testing.T) {
		user, err := userService.GetUser("user1")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "AliceWonder", user.Username)
		assert.Equal(t, "alice.wonder@example.com", user.Email)
	})

	// Test GetUserBasicsByPrefix
	t.Run("GetUserBasicsByPrefix", func(t *testing.T) {
		users, err := userService.GetUserBasicsByPrefix("user")
		assert.NoError(t, err)
		assert.NotEmpty(t, users)
		assert.GreaterOrEqual(t, len(users), 3)
		for _, user := range users {
			assert.Contains(t, string(user.UserID), "user")
		}
	})

	// Test GetGameStatByIDs
	t.Run("GetGameStat", func(t *testing.T) {
		gameStat, err := userService.GetGameStat("user1", "soccer")
		assert.NoError(t, err)
		assert.NotNil(t, gameStat)
		assert.Equal(t, models.UserID("user1"), gameStat.UserID)
		assert.Equal(t, models.GameID("soccer"), gameStat.GameID)
		assert.Equal(t, models.AttributeStat(1), gameStat.GameAttributes["goals"])
	})

	// Test CreateUser
	t.Run("CreateUser", func(t *testing.T) {
		newUser, err := models.NewUser("user6", "TestUser", "test@example.com", []models.GameID{})
		assert.NoError(t, err)
		createdUser, err := userService.CreateUser(newUser)
		assert.NoError(t, err)
		assert.NotNil(t, createdUser)
		assert.Equal(t, newUser.Username, createdUser.Username)
		assert.Equal(t, newUser.Email, createdUser.Email)

		// Clean up: Delete the created user
		err = userService.DeleteUser(&createdUser.UserID)
		assert.NoError(t, err)
	})

	// Test UpdateUser
	t.Run("UpdateUser", func(t *testing.T) {
		user, err := userService.GetUser("user2")
		assert.NoError(t, err)

		originalUsername := user.Username
		originalEmail := user.Email

		updatedUser, err := models.NewUser(user.UserID, "UpdatedUser", "updated@example.com", user.GamesPlayed)
		assert.NoError(t, err)

		result, err := userService.UpdateUser(updatedUser)
		assert.NoError(t, err)
		assert.Equal(t, updatedUser, result)

		// Clean up: Revert the user to original state
		revertedUser, err := models.NewUser(user.UserID, originalUsername, originalEmail, user.GamesPlayed)
		assert.NoError(t, err)
		_, err = userService.UpdateUser(revertedUser)
		assert.NoError(t, err)
	})

	// Test DeleteUser
	t.Run("DeleteUser", func(t *testing.T) {
		newUser, err := models.NewUser("tempuser", "TempUser", "temp@example.com", []models.GameID{})
		assert.NoError(t, err)
		createdUser, err := userService.CreateUser(newUser)
		assert.NoError(t, err)

		err = userService.DeleteUser(&createdUser.UserID)
		assert.NoError(t, err)

		// Verify user is deleted
		_, err = userService.GetUser(createdUser.UserID)
		assert.Error(t, err)
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
