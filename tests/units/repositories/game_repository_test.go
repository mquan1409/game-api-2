package tests

import (
	"testing"

	"github.com/mquan1409/game-api/internal/config"
	"github.com/mquan1409/game-api/internal/models"
	"github.com/mquan1409/game-api/internal/repositories"
	"github.com/mquan1409/game-api/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func TestGameRepository(t *testing.T) {
	// Load test configuration
	cfg := config.LoadConfig("development")

	// Setup
	db, err := utils.SetupTestDB(&cfg)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	repo := repositories.NewDynamoDBGameRepository(db, cfg.TableName)

	// Scan the entire table before tests
	beforeScan, err := utils.ScanEntireTable(db, cfg.TableName)
	if err != nil {
		t.Fatalf("Failed to scan table before tests: %v", err)
	}

	// Test GetGame
	t.Run("GetGame", func(t *testing.T) {
		game, err := repo.GetGame("soccer")
		assert.NoError(t, err)
		assert.NotEmpty(t, game)
		assert.Equal(t, "Soccer", game.Description)
		assert.ElementsMatch(t, []models.AttributeName{"elo", "goals", "assists", "shots_on_target", "passes_completed"}, game.Attributes)
		assert.ElementsMatch(t, []models.AttributeName{"elo"}, game.RankedAttributes)
	})

	// Test CreateGame
	t.Run("CreateGame", func(t *testing.T) {
		newGame, err := models.NewGame("basketball", "Basketball", []models.AttributeName{"elo", "points", "rebounds", "assists", "steals"}, []models.AttributeName{"elo"})
		assert.NoError(t, err)
		createdGame, err := repo.CreateGame(newGame, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdGame)
		assert.Equal(t, newGame.Description, createdGame.Description)
		assert.ElementsMatch(t, newGame.Attributes, createdGame.Attributes)
		assert.ElementsMatch(t, newGame.RankedAttributes, createdGame.RankedAttributes)

		// Verify the game was actually created
		fetchedGame, err := repo.GetGame(createdGame.GameID)
		assert.NoError(t, err)
		assert.Equal(t, createdGame, fetchedGame)

		// Clean up: delete the created game
		err = repo.DeleteGame("basketball", nil)
		assert.NoError(t, err)
	})

	// Test UpdateGame
	t.Run("UpdateGame", func(t *testing.T) {
		game, err := repo.GetGame("pool")
		assert.NoError(t, err)

		updatedGame, err := models.NewGame(game.GameID, "Updated Pool", append(game.Attributes, "trick_shots"), game.RankedAttributes)
		assert.NoError(t, err)

		result, err := repo.UpdateGame(updatedGame, nil)
		assert.NoError(t, err)
		assert.Equal(t, updatedGame, result)

		// Verify the game was actually updated
		fetchedGame, err := repo.GetGame(game.GameID)
		assert.NoError(t, err)
		assert.Equal(t, updatedGame, fetchedGame)

		// Clean up: revert the game to its original state
		_, err = repo.UpdateGame(game, nil)
		assert.NoError(t, err)
	})

	// Test DeleteGame
	t.Run("DeleteGame", func(t *testing.T) {
		// First, create a game to delete
		tempGame, err := models.NewGame("temp_game", "Temporary Game", []models.AttributeName{"elo", "temp_attr1", "temp_attr2"}, []models.AttributeName{"elo"})
		assert.NoError(t, err)
		_, err = repo.CreateGame(tempGame, nil)
		assert.NoError(t, err)

		// Now delete the game
		err = repo.DeleteGame("temp_game", nil)
		assert.NoError(t, err)

		// Verify game is deleted
		_, err = repo.GetGame("temp_game")
		assert.Error(t, err)
	})

	// Test CreateAndUpdateInTransaction
	t.Run("CreateAndUpdateInTransaction", func(t *testing.T) {
		tx := &dynamodb.TransactWriteItemsInput{}

		newGame, err := models.NewGame("tennis", "Tennis", []models.AttributeName{"elo", "aces", "double_faults", "winners", "unforced_errors"}, []models.AttributeName{"elo"})
		assert.NoError(t, err)
		createdGame, err := repo.CreateGame(newGame, tx)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdGame)

		existingGame, err := repo.GetGame("soccer")
		assert.NoError(t, err)

		updatedGame, err := models.NewGame(existingGame.GameID, "Updated Soccer", append(existingGame.Attributes, "tackles"), existingGame.RankedAttributes)
		assert.NoError(t, err)

		result, err := repo.UpdateGame(updatedGame, tx)
		assert.NoError(t, err)
		assert.Equal(t, updatedGame, result)

		_, err = db.TransactWriteItems(tx)
		assert.NoError(t, err)

		// Verify the new game was created
		fetchedNewGame, err := repo.GetGame("tennis")
		assert.NoError(t, err)
		assert.Equal(t, newGame, fetchedNewGame)

		// Verify the existing game was updated
		fetchedUpdatedGame, err := repo.GetGame("soccer")
		assert.NoError(t, err)
		assert.Equal(t, updatedGame, fetchedUpdatedGame)

		// Clean up: delete the new game and revert the updated game
		err = repo.DeleteGame("tennis", nil)
		assert.NoError(t, err)

		_, err = repo.UpdateGame(existingGame, nil)
		assert.NoError(t, err)
	})

	// Test DeleteAndUpdateInTransaction
	t.Run("DeleteAndUpdateInTransaction", func(t *testing.T) {
		tx := &dynamodb.TransactWriteItemsInput{}

		// First, create a game to delete
		tempGame, err := models.NewGame("temp_game", "Temporary Game", []models.AttributeName{"elo", "temp_attr1", "temp_attr2"}, []models.AttributeName{"elo"})
		assert.NoError(t, err)
		_, err = repo.CreateGame(tempGame, nil)
		assert.NoError(t, err)

		// Delete the temporary game
		err = repo.DeleteGame("temp_game", tx)
		assert.NoError(t, err)

		// Update an existing game
		existingGame, err := repo.GetGame("pickleball")
		assert.NoError(t, err)

		updatedGame, err := models.NewGame(existingGame.GameID, "Updated Pickleball", append(existingGame.Attributes, "lobs"), existingGame.RankedAttributes)
		assert.NoError(t, err)

		result, err := repo.UpdateGame(updatedGame, tx)
		assert.NoError(t, err)
		assert.Equal(t, updatedGame, result)

		_, err = db.TransactWriteItems(tx)
		assert.NoError(t, err)

		// Verify the temp game was deleted
		_, err = repo.GetGame("temp_game")
		assert.Error(t, err)

		// Verify the existing game was updated
		fetchedUpdatedGame, err := repo.GetGame("pickleball")
		assert.NoError(t, err)
		assert.Equal(t, updatedGame, fetchedUpdatedGame)

		// Clean up: revert the updated game
		_, err = repo.UpdateGame(existingGame, nil)
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
