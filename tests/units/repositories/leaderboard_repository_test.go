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

func TestLeaderboardRepository(t *testing.T) {
	// Load test configuration
	cfg := config.LoadConfig("development")

	// Setup
	db, err := utils.SetupTestDB(&cfg)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	repo := repositories.NewDynamoDBLeaderboardRepository(db, cfg.TableName)

	// Scan the entire table before tests
	beforeScan, err := utils.ScanEntireTable(db, cfg.TableName)
	if err != nil {
		t.Fatalf("Failed to scan table before tests: %v", err)
	}

	// Test GetLeaderboard
	t.Run("GetLeaderboard", func(t *testing.T) {
		leaderboard, err := repo.GetLeaderboard(models.GameID("soccer"), models.AttributeName("elo"))
		assert.NoError(t, err)
		assert.NotNil(t, leaderboard)
		assert.Equal(t, 5, len(leaderboard.UserIDs))
		assert.Equal(t, models.UserID("user2"), leaderboard.UserIDs[0])
		assert.Equal(t, models.UserID("user1"), leaderboard.UserIDs[1])
		assert.Equal(t, models.UserID("user3"), leaderboard.UserIDs[2])
		assert.Equal(t, models.UserID("dianadancer"), leaderboard.UserIDs[3])
		assert.Equal(t, models.UserID("eveexplorer"), leaderboard.UserIDs[4])
	})

	// Test GetBoundedLeaderboard
	t.Run("GetBoundedLeaderboard", func(t *testing.T) {
		boundedLeaderboard, err := repo.GetBoundedLeaderboard(models.GameID("soccer"), models.AttributeName("elo"), 3)
		assert.NoError(t, err)
		assert.NotNil(t, boundedLeaderboard)
		assert.Equal(t, 3, len(boundedLeaderboard.UserIDs))
		assert.Equal(t, models.UserID("user2"), boundedLeaderboard.UserIDs[0])
		assert.Equal(t, models.UserID("user1"), boundedLeaderboard.UserIDs[1])
		assert.Equal(t, models.UserID("user3"), boundedLeaderboard.UserIDs[2])
	})

	// Test AddLeaderboardItem
	t.Run("AddLeaderboardItem", func(t *testing.T) {
		err := repo.AddLeaderboardItem(models.GameID("soccer"), models.UserID("newuser"), models.AttributeName("elo"), models.AttributeStat(5), nil)
		assert.NoError(t, err)

		// Verify the leaderboard was updated
		leaderboard, err := repo.GetLeaderboard(models.GameID("soccer"), models.AttributeName("elo"))
		assert.NoError(t, err)
		assert.Equal(t, 6, len(leaderboard.UserIDs))
		assert.Equal(t, models.UserID("newuser"), leaderboard.UserIDs[0])

		// Clean up
		err = repo.DeleteLeaderboardItem(models.GameID("soccer"), models.UserID("newuser"), models.AttributeName("elo"), models.AttributeStat(5), nil)
		assert.NoError(t, err)
	})

	// Test UpdateLeaderboardItem
	t.Run("UpdateLeaderboardItem", func(t *testing.T) {
		oldLeaderboard, err := repo.GetLeaderboard(models.GameID("soccer"), models.AttributeName("elo"))
		assert.NoError(t, err)
		t.Logf("Old leaderboard: %v", oldLeaderboard)

		err = repo.UpdateLeaderboardItem(models.GameID("soccer"), models.UserID("user1"), models.AttributeName("elo"), models.AttributeStat(10), models.AttributeStat(2), nil)
		assert.NoError(t, err)

		// Verify the leaderboard was updated
		newLeaderboard, err := repo.GetLeaderboard(models.GameID("soccer"), models.AttributeName("elo"))
		assert.NoError(t, err)
		t.Logf("New leaderboard: %v", newLeaderboard)

		assert.Equal(t, models.UserID("user1"), newLeaderboard.UserIDs[0], "Expected user1 to be at the top after update")
		assert.NotEqual(t, oldLeaderboard.UserIDs[0], newLeaderboard.UserIDs[0], "Expected top user to change after update")

		// Revert changes
		err = repo.UpdateLeaderboardItem(models.GameID("soccer"), models.UserID("user1"), models.AttributeName("elo"), models.AttributeStat(2), models.AttributeStat(10), nil)
		assert.NoError(t, err)
	})

	// Test DeleteLeaderboardItem
	t.Run("DeleteLeaderboardItem", func(t *testing.T) {
		// First, add a new item to delete
		err := repo.AddLeaderboardItem(models.GameID("deletegame"), models.UserID("deleteuser"), models.AttributeName("elo"), models.AttributeStat(5), nil)
		assert.NoError(t, err)

		// Now delete it
		err = repo.DeleteLeaderboardItem(models.GameID("deletegame"), models.UserID("deleteuser"), models.AttributeName("elo"), models.AttributeStat(5), nil)
		assert.NoError(t, err)

		// Verify the leaderboard for this game is empty
		leaderboard, err := repo.GetLeaderboard(models.GameID("deletegame"), models.AttributeName("elo"))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(leaderboard.UserIDs))
	})

	// Test AddLeaderboardItem and UpdateLeaderboardItem in the same transaction
	t.Run("AddAndUpdateLeaderboardInTransaction", func(t *testing.T) {
		tx := &dynamodb.TransactWriteItemsInput{}

		// Add new leaderboard item
		err := repo.AddLeaderboardItem(models.GameID("soccer"), models.UserID("transactionuser"), models.AttributeName("elo"), models.AttributeStat(7), tx)
		assert.NoError(t, err)

		// Update existing leaderboard item
		err = repo.UpdateLeaderboardItem(models.GameID("soccer"), models.UserID("user1"), models.AttributeName("elo"), models.AttributeStat(8), models.AttributeStat(2), tx)
		assert.NoError(t, err)

		_, err = db.TransactWriteItems(tx)
		assert.NoError(t, err)

		// Verify the leaderboard was updated
		leaderboard, err := repo.GetLeaderboard(models.GameID("soccer"), models.AttributeName("elo"))
		assert.NoError(t, err)
		assert.Equal(t, 6, len(leaderboard.UserIDs))
		assert.Equal(t, models.UserID("user1"), leaderboard.UserIDs[0])
		assert.Equal(t, models.UserID("transactionuser"), leaderboard.UserIDs[1])

		// Clean up
		cleanupTx := &dynamodb.TransactWriteItemsInput{}

		err = repo.DeleteLeaderboardItem(models.GameID("soccer"), models.UserID("transactionuser"), models.AttributeName("elo"), models.AttributeStat(7), cleanupTx)
		assert.NoError(t, err)

		err = repo.UpdateLeaderboardItem(models.GameID("soccer"), models.UserID("user1"), models.AttributeName("elo"), models.AttributeStat(2), models.AttributeStat(8), cleanupTx)
		assert.NoError(t, err)

		_, err = db.TransactWriteItems(cleanupTx)
		assert.NoError(t, err)
	})

	// Test DeleteLeaderboardItemsByGameAndUser
	t.Run("DeleteLeaderboardItemsByGameAndUser", func(t *testing.T) {
		// Add some items for a new game and user
		err := repo.AddLeaderboardItem(models.GameID("testgame"), models.UserID("testuser"), models.AttributeName("elo"), models.AttributeStat(10), nil)
		assert.NoError(t, err)
		err = repo.AddLeaderboardItem(models.GameID("testgame"), models.UserID("testuser"), models.AttributeName("score"), models.AttributeStat(100), nil)
		assert.NoError(t, err)

		// Delete all items for this game and user
		err = repo.DeleteLeaderboardItemsByGameAndUser(models.GameID("testgame"), models.UserID("testuser"), nil)
		assert.NoError(t, err)

		// Verify the items were deleted
		leaderboard, err := repo.GetLeaderboard(models.GameID("testgame"), models.AttributeName("elo"))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(leaderboard.UserIDs))

		leaderboard, err = repo.GetLeaderboard(models.GameID("testgame"), models.AttributeName("score"))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(leaderboard.UserIDs))
	})

	// Test DeleteLeaderboardItemsByGame
	t.Run("DeleteLeaderboardItemsByGame", func(t *testing.T) {
		// Add some items for a new game
		err := repo.AddLeaderboardItem(models.GameID("deletegame"), models.UserID("user1"), models.AttributeName("elo"), models.AttributeStat(10), nil)
		assert.NoError(t, err)
		err = repo.AddLeaderboardItem(models.GameID("deletegame"), models.UserID("user2"), models.AttributeName("elo"), models.AttributeStat(20), nil)
		assert.NoError(t, err)

		// Delete all items for this game
		err = repo.DeleteLeaderboardItemsByGame(models.GameID("deletegame"), nil)
		assert.NoError(t, err)

		// Verify the items were deleted
		leaderboard, err := repo.GetLeaderboard(models.GameID("deletegame"), models.AttributeName("elo"))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(leaderboard.UserIDs))
	})

	// Test DeleteLeaderboardItemsByGameAndAttribute
	t.Run("DeleteLeaderboardItemsByGameAndAttribute", func(t *testing.T) {
		// Add some items for a new game with different attributes
		err := repo.AddLeaderboardItem(models.GameID("attributegame"), models.UserID("user1"), models.AttributeName("elo"), models.AttributeStat(10), nil)
		assert.NoError(t, err)
		err = repo.AddLeaderboardItem(models.GameID("attributegame"), models.UserID("user2"), models.AttributeName("elo"), models.AttributeStat(20), nil)
		assert.NoError(t, err)
		err = repo.AddLeaderboardItem(models.GameID("attributegame"), models.UserID("user1"), models.AttributeName("score"), models.AttributeStat(100), nil)
		assert.NoError(t, err)
		err = repo.AddLeaderboardItem(models.GameID("attributegame"), models.UserID("user2"), models.AttributeName("score"), models.AttributeStat(200), nil)
		assert.NoError(t, err)

		// Delete all items for this game and the "elo" attribute
		err = repo.DeleteLeaderboardItemsByGameAndAttribute(models.GameID("attributegame"), models.AttributeName("elo"), nil)
		assert.NoError(t, err)

		// Verify the "elo" items were deleted
		leaderboard, err := repo.GetLeaderboard(models.GameID("attributegame"), models.AttributeName("elo"))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(leaderboard.UserIDs))

		// Verify the "score" items still exist
		leaderboard, err = repo.GetLeaderboard(models.GameID("attributegame"), models.AttributeName("score"))
		assert.NoError(t, err)
		assert.Equal(t, 2, len(leaderboard.UserIDs))

		// Clean up
		err = repo.DeleteLeaderboardItemsByGame(models.GameID("attributegame"), nil)
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
