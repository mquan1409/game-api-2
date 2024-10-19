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

func TestGameStatRepository(t *testing.T) {
	// Load test configuration
	cfg := config.LoadConfig("development")

	// Setup
	db, err := utils.SetupTestDB(&cfg)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	repo := repositories.NewDynamoDBGameStatRepository(db, cfg.TableName)

	// Scan the entire table before tests
	beforeScan, err := utils.ScanEntireTable(db, cfg.TableName)
	if err != nil {
		t.Fatalf("Failed to scan table before tests: %v", err)
	}

	// Test GetGameStat
	t.Run("GetGameStat", func(t *testing.T) {
		gameStat, err := repo.GetGameStat(models.UserID("user1"), models.GameID("soccer"))
		assert.NoError(t, err)
		assert.NotNil(t, gameStat)
		assert.Equal(t, models.UserID("user1"), gameStat.UserID)
		assert.Equal(t, models.GameID("soccer"), gameStat.GameID)
		assert.Equal(t, models.AttributeStat(1), gameStat.GameAttributes["goals"])
		assert.Equal(t, models.AttributeStat(1), gameStat.GameAttributes["assists"])
		assert.Equal(t, models.AttributeStat(3), gameStat.GameAttributes["shots_on_target"])
		assert.Equal(t, models.AttributeStat(20), gameStat.GameAttributes["passes_completed"])
	})

	// Test CreateGameStat
	t.Run("CreateGameStat", func(t *testing.T) {
		failedAttributes := models.AttributesStatsMap{
			"goals":             1,
			"assists":           -1,
			"shots_on_target":   0,
			"passes_completed":  -1,
		}
		_, err := models.NewGameStat(models.UserID("newuser"), models.GameID("soccer"), failedAttributes)
		assert.Error(t, err)

		attributes := models.AttributesStatsMap{
			"goals":             1,
			"assists":           0,
			"shots_on_target":   2,
			"passes_completed":  15,
		}
		newGameStat, err := models.NewGameStat(models.UserID("newuser"), models.GameID("soccer"), attributes)
		assert.NoError(t, err)

		err = repo.CreateGameStat(newGameStat, nil)
		assert.NoError(t, err)

		// Verify the game stat was actually created
		fetchedGameStat, err := repo.GetGameStat(models.UserID("newuser"), models.GameID("soccer"))
		assert.NoError(t, err)
		assert.Equal(t, newGameStat, fetchedGameStat)

		// Clean up: delete the created game stat
		err = repo.DeleteGameStat(models.UserID("newuser"), models.GameID("soccer"), nil)
		assert.NoError(t, err)
	})

	// Test UpdateGameStat
	t.Run("UpdateGameStat", func(t *testing.T) {
		oldGameStat, err := repo.GetGameStat(models.UserID("user1"), models.GameID("soccer"))
		assert.NoError(t, err)

		attributes := models.AttributesStatsMap{
			"goals":             2,
			"assists":           1,
			"shots_on_target":   4,
			"passes_completed":  25,
		}
		updatedGameStat, err := models.NewGameStat(models.UserID("user1"), models.GameID("soccer"), attributes)
		assert.NoError(t, err)

		err = repo.UpdateGameStat(updatedGameStat, nil)
		assert.NoError(t, err)

		// Verify the game stat was actually updated
		fetchedGameStat, err := repo.GetGameStat(models.UserID("user1"), models.GameID("soccer"))
		assert.NoError(t, err)
		assert.Equal(t, updatedGameStat, fetchedGameStat)
		assert.NotEqual(t, oldGameStat, fetchedGameStat)

		// Clean up: revert the game stat to its original state
		err = repo.UpdateGameStat(oldGameStat, nil)
		assert.NoError(t, err)
	})

	// Test DeleteGameStat
	t.Run("DeleteGameStat", func(t *testing.T) {
		// Create a temporary game stat to delete
		tempAttributes := models.AttributesStatsMap{
			"goals":             1,
			"assists":           1,
			"shots_on_target":   1,
			"passes_completed":  10,
		}
		tempGameStat, err := models.NewGameStat(models.UserID("tempuser"), models.GameID("soccer"), tempAttributes)
		assert.NoError(t, err)
		err = repo.CreateGameStat(tempGameStat, nil)
		assert.NoError(t, err)

		err = repo.DeleteGameStat(models.UserID("tempuser"), models.GameID("soccer"), nil)
		assert.NoError(t, err)

		// Verify game stat is deleted
		_, err = repo.GetGameStat(models.UserID("tempuser"), models.GameID("soccer"))
		assert.Error(t, err)
	})

	// Test CreateGameStat and UpdateGameStat in the same transaction
	t.Run("CreateAndUpdateGameStatInTransaction", func(t *testing.T) {
		tx := &dynamodb.TransactWriteItemsInput{}
		oldGameStat, err := repo.GetGameStat(models.UserID("user1"), models.GameID("soccer"))
		assert.NoError(t, err)

		newAttributes := models.AttributesStatsMap{
			"goals":             1,
			"assists":           1,
			"shots_on_target":   2,
			"passes_completed":  18,
		}
		newGameStat, err := models.NewGameStat(models.UserID("transactionuser"), models.GameID("soccer"), newAttributes)
		assert.NoError(t, err)

		err = repo.CreateGameStat(newGameStat, tx)
		assert.NoError(t, err)

		updatedAttributes := models.AttributesStatsMap{
			"goals":             3,
			"assists":           2,
			"shots_on_target":   5,
			"passes_completed":  30,
		}
		updatedGameStat, err := models.NewGameStat(models.UserID("user1"), models.GameID("soccer"), updatedAttributes)
		assert.NoError(t, err)

		err = repo.UpdateGameStat(updatedGameStat, tx)
		assert.NoError(t, err)

		_, err = db.TransactWriteItems(tx)
		assert.NoError(t, err)

		// Verify the new game stat was created
		createdGameStat, err := repo.GetGameStat(models.UserID("transactionuser"), models.GameID("soccer"))
		assert.NoError(t, err)
		assert.Equal(t, newGameStat, createdGameStat)

		// Verify the existing game stat was updated
		fetchedUpdatedGameStat, err := repo.GetGameStat(models.UserID("user1"), models.GameID("soccer"))
		assert.NoError(t, err)
		assert.Equal(t, updatedGameStat, fetchedUpdatedGameStat)
		assert.NotEqual(t, oldGameStat, fetchedUpdatedGameStat)

		// Clean up
		cleanupTx := &dynamodb.TransactWriteItemsInput{}
		err = repo.DeleteGameStat(models.UserID("transactionuser"), models.GameID("soccer"), cleanupTx)
		assert.NoError(t, err)
		err = repo.UpdateGameStat(oldGameStat, cleanupTx)
		assert.NoError(t, err)
		_, err = db.TransactWriteItems(cleanupTx)
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
