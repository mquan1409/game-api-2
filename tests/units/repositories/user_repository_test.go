package tests

import (
	"testing"
	"github.com/mquan1409/game-api/internal/models"
	"github.com/mquan1409/game-api/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/mquan1409/game-api/internal/config"
	"github.com/mquan1409/game-api/internal/utils"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func TestUserRepository(t *testing.T) {
	// Load test configuration
	cfg := config.LoadConfig("development")

	// Setup
	db, err := utils.SetupTestDB(&cfg)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	repo := repositories.NewDynamoDBUserRepository(db, cfg.TableName)

	// Scan the entire table before tests
	beforeScan, err := utils.ScanEntireTable(db, cfg.TableName)
	if err != nil {
		t.Fatalf("Failed to scan table before tests: %v", err)
	}

	// Test GetUser
	t.Run("GetUser", func(t *testing.T) {
		user, err := repo.GetUser("user1")
		assert.NoError(t, err)
		assert.NotEmpty(t, user)
		assert.Equal(t, "AliceWonder", user.Username)
		assert.Equal(t, "alice.wonder@example.com", user.Email)
	})

	// Test GetUsersByPrefix
	t.Run("GetUsersByPrefix", func(t *testing.T) {
		// Test with prefix "user"
		userPrefixUsers, err := repo.GetUserBasicsByPrefix("user")
		assert.NoError(t, err)
		assert.NotEmpty(t, userPrefixUsers)
		assert.GreaterOrEqual(t, len(userPrefixUsers), 3)
		for _, user := range userPrefixUsers {
			assert.Contains(t, string(user.UserID), "user")
		}

		// Test with a prefix that should return no results
		emptyUsers, err := repo.GetUserBasicsByPrefix("NonexistentPrefix")
		assert.NoError(t, err)
		assert.Empty(t, emptyUsers)
	})

	// Test CreateUser
	t.Run("CreateUser", func(t *testing.T) {
		newUser, err := models.NewUser("user6", "TestUser", "test@example.com", []models.GameID{})
		assert.NoError(t, err)
		createdUser, err := repo.CreateUser(newUser, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdUser)
		assert.Equal(t, newUser.Username, createdUser.Username)
		assert.Equal(t, newUser.Email, createdUser.Email)

		// Verify the user was actually created
		fetchedUser, err := repo.GetUser(createdUser.UserID)
		assert.NoError(t, err)
		assert.Equal(t, createdUser, fetchedUser)

		// Clean up: Delete the created user
		err = repo.DeleteUser(&createdUser.UserID, nil)
		assert.NoError(t, err)
	})

	// Test UpdateUser
	t.Run("UpdateUser", func(t *testing.T) {
		user, err := repo.GetUser("user2")
		assert.NoError(t, err)

		originalUsername := user.Username
		originalEmail := user.Email

		updatedUser, err := models.NewUser(user.UserID, "UpdatedUser", "updated@example.com", user.GamesPlayed)
		assert.NoError(t, err)

		result, err := repo.UpdateUser(updatedUser, nil)
		assert.NoError(t, err)
		assert.Equal(t, updatedUser, result)

		// Verify the user was actually updated
		fetchedUser, err := repo.GetUser(user.UserID)
		assert.NoError(t, err)
		assert.Equal(t, updatedUser, fetchedUser)

		// Clean up: Revert the user to original state
		revertedUser, err := models.NewUser(user.UserID, originalUsername, originalEmail, user.GamesPlayed)
		assert.NoError(t, err)
		_, err = repo.UpdateUser(revertedUser, nil)
		assert.NoError(t, err)
	})

	// Test DeleteUser
	t.Run("DeleteUser", func(t *testing.T) {
		newUser, err := models.NewUser("tempuser", "TempUser", "temp@example.com", []models.GameID{})
		assert.NoError(t, err)
		createdUser, err := repo.CreateUser(newUser, nil)
		assert.NoError(t, err)

		err = repo.DeleteUser(&createdUser.UserID, nil)
		assert.NoError(t, err)

		// Verify user is deleted
		_, err = repo.GetUser(createdUser.UserID)
		assert.Error(t, err)
	})

	// Test CreateUser and UpdateUser in the same transaction
	t.Run("CreateAndUpdateUserInTransaction", func(t *testing.T) {
		tx := &dynamodb.TransactWriteItemsInput{}

		newUser, err := models.NewUser("user7", "NewUser", "new@example.com", []models.GameID{"soccer"})
		assert.NoError(t, err)

		_, err = repo.CreateUser(newUser, tx)
		assert.NoError(t, err)

		existingUser, err := repo.GetUser("user3")
		assert.NoError(t, err)

		originalUsername := existingUser.Username
		originalGamesPlayed := existingUser.GamesPlayed

		updatedUser, err := models.NewUser(existingUser.UserID, "UpdatedCharlie", existingUser.Email, append(existingUser.GamesPlayed, "pickleball"))
		assert.NoError(t, err)

		_, err = repo.UpdateUser(updatedUser, tx)
		assert.NoError(t, err)

		_, err = db.TransactWriteItems(tx)
		assert.NoError(t, err)

		// Verify the new user was created
		createdUser, err := repo.GetUser("user7")
		assert.NoError(t, err)
		assert.Equal(t, newUser, createdUser)

		// Verify the existing user was updated
		fetchedUpdatedUser, err := repo.GetUser("user3")
		assert.NoError(t, err)
		assert.Equal(t, updatedUser, fetchedUpdatedUser)

		// Clean up: Delete the created user and revert the updated user
		cleanupTx := &dynamodb.TransactWriteItemsInput{}

		err = repo.DeleteUser(&newUser.UserID, cleanupTx)
		assert.NoError(t, err)

		revertedUser, err := models.NewUser(existingUser.UserID, originalUsername, existingUser.Email, originalGamesPlayed)
		assert.NoError(t, err)
		_, err = repo.UpdateUser(revertedUser, cleanupTx)
		assert.NoError(t, err)

		_, err = db.TransactWriteItems(cleanupTx)
		assert.NoError(t, err)
	})

	// Test UpdateUser and DeleteUser in the same transaction, followed by a separate CreateUser
	t.Run("UpdateDeleteInTransaction", func(t *testing.T) {
		// First, create users to work with
		initialUser, err := models.NewUser("user8", "InitialUser", "initial@example.com", []models.GameID{"tennis"})
		assert.NoError(t, err)
		_, err = repo.CreateUser(initialUser, nil)
		assert.NoError(t, err)

		second_initialUser, err := models.NewUser("user9", "InitialUser", "initial@example.com", []models.GameID{"tennis"})
		assert.NoError(t, err)
		_, err = repo.CreateUser(second_initialUser, nil)
		assert.NoError(t, err)

		// Start a new transaction
		tx := &dynamodb.TransactWriteItemsInput{}

		// Update the user
		updatedUser, err := models.NewUser("user8", "UpdatedUser", "updated@example.com", []models.GameID{"tennis", "badminton"})
		assert.NoError(t, err)
		_, err = repo.UpdateUser(updatedUser, tx)
		assert.NoError(t, err)

		// Delete the user in the same transaction
		user9ID := models.UserID("user9")
		err = repo.DeleteUser(&user9ID, tx)
		assert.NoError(t, err)

		// Execute the transaction
		_, err = db.TransactWriteItems(tx)
		assert.NoError(t, err)

		// Verify the user was deleted
		_, err = repo.GetUser("user9")
		assert.Error(t, err)

		fetchedUser, err := repo.GetUser("user8")
		assert.NoError(t, err)
		assert.Equal(t, fetchedUser.Username, "UpdatedUser")	
		assert.Equal(t, fetchedUser.Email, "updated@example.com")
		assert.Equal(t, fetchedUser.GamesPlayed, []models.GameID{"tennis", "badminton"})

		// Clean up: Delete the remaining user
		user8ID := models.UserID("user8")
		err = repo.DeleteUser(&user8ID, nil)
		assert.NoError(t, err)
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
