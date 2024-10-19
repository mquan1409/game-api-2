package tests

import (
	"testing"

	"github.com/mquan1409/game-api/internal/config"
	"github.com/mquan1409/game-api/internal/models"
	"github.com/mquan1409/game-api/internal/repositories"
	"github.com/mquan1409/game-api/internal/services"
	"github.com/mquan1409/game-api/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestGameService(t *testing.T) {
	// Load test configuration
	cfg := config.LoadConfig("development")

	// Setup
	db, err := utils.SetupTestDB(&cfg)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	gameRepo := repositories.NewDynamoDBGameRepository(db, cfg.TableName)
	leaderboardRepo := repositories.NewDynamoDBLeaderboardRepository(db, cfg.TableName)
	gameService := services.NewGameServiceImpl(gameRepo, leaderboardRepo)

	// Scan the entire table before tests
	beforeScan, err := utils.ScanEntireTable(db, cfg.TableName)
	if err != nil {
		t.Fatalf("Failed to scan table before tests: %v", err)
	}

	// Test GetGame
	t.Run("GetGame", func(t *testing.T) {
		game, err := gameService.GetGame("soccer")
		assert.NoError(t, err)
		assert.NotNil(t, game)
		assert.Equal(t, models.GameID("soccer"), game.GameID)
		assert.Equal(t, "Soccer", game.Description)
		assert.ElementsMatch(t, []models.AttributeName{"elo", "goals", "assists", "shots_on_target", "passes_completed"}, game.Attributes)
		assert.ElementsMatch(t, []models.AttributeName{"elo"}, game.RankedAttributes)
	})

	// Test GetGameLeaderboard
	t.Run("GetGameLeaderboard", func(t *testing.T) {
		leaderboard, err := gameService.GetGameLeaderboard("soccer", "elo")
		assert.NoError(t, err)
		assert.NotEmpty(t, leaderboard)
		assert.Equal(t, 5, len(leaderboard))
		assert.Equal(t, models.UserID("user2"), leaderboard[0])
		assert.Equal(t, models.UserID("user1"), leaderboard[1])
		assert.Equal(t, models.UserID("user3"), leaderboard[2])
		assert.Equal(t, models.UserID("dianadancer"), leaderboard[3])
		assert.Equal(t, models.UserID("eveexplorer"), leaderboard[4])
	})

	// Test GetBoundedGameLeaderboard
	t.Run("GetBoundedGameLeaderboard", func(t *testing.T) {
		boundedLeaderboard, err := gameService.GetBoundedGameLeaderboard("soccer", "elo", 3)
		assert.NoError(t, err)
		assert.NotEmpty(t, boundedLeaderboard)
		assert.Equal(t, 3, len(boundedLeaderboard))
		assert.Equal(t, models.UserID("user2"), boundedLeaderboard[0])
		assert.Equal(t, models.UserID("user1"), boundedLeaderboard[1])
		assert.Equal(t, models.UserID("user3"), boundedLeaderboard[2])
	})

	// Test CreateGame
	t.Run("CreateGame", func(t *testing.T) {
		newGame, err := models.NewGame("testgame", "A test game", []models.AttributeName{"score"}, []models.AttributeName{"score"})
		assert.NoError(t, err)
		createdGame, err := gameService.CreateGame(newGame)
		assert.NoError(t, err)
		assert.NotNil(t, createdGame)
		assert.Equal(t, newGame.GameID, createdGame.GameID)
		assert.Equal(t, newGame.Description, createdGame.Description)

		// Verify the game was created
		retrievedGame, err := gameService.GetGame("testgame")
		assert.NoError(t, err)
		assert.Equal(t, createdGame, retrievedGame)

		// Clean up: Delete the created game
		err = gameService.DeleteGame(createdGame.GameID)
		assert.NoError(t, err)
	})

	// Test UpdateGame
	t.Run("UpdateGame", func(t *testing.T) {
		// Create a new game for testing updates
		newGame, err := models.NewGame("updategame", "Game for update tests", []models.AttributeName{"score", "time", "level"}, []models.AttributeName{"score", "time"})
		assert.NoError(t, err)
		createdGame, err := gameService.CreateGame(newGame)
		assert.NoError(t, err)

		// Test case 1: Update description
		t.Run("UpdateDescription", func(t *testing.T) {
			updatedGame, err := models.NewGame(createdGame.GameID, "Updated game description", createdGame.Attributes, createdGame.RankedAttributes)
			assert.NoError(t, err)

			result, err := gameService.UpdateGame(updatedGame)
			assert.NoError(t, err)
			assert.Equal(t, updatedGame, result)

			// Verify the game was updated
			retrievedGame, err := gameService.GetGame(createdGame.GameID)
			assert.NoError(t, err)
			assert.Equal(t, "Updated game description", retrievedGame.Description)
		})

		// Test case 2: Add a new ranked attribute
		t.Run("AddRankedAttribute", func(t *testing.T) {
			newRankedAttributes := append(createdGame.RankedAttributes, models.AttributeName("level"))
			updatedGame, err := models.NewGame(createdGame.GameID, createdGame.Description, createdGame.Attributes, newRankedAttributes)
			assert.NoError(t, err)

			result, err := gameService.UpdateGame(updatedGame)
			assert.NoError(t, err)
			assert.Equal(t, updatedGame, result)

			// Verify the game was updated
			retrievedGame, err := gameService.GetGame(createdGame.GameID)
			assert.NoError(t, err)
			assert.Contains(t, retrievedGame.RankedAttributes, models.AttributeName("level"))
		})

		// Test case 3: Remove a ranked attribute
		t.Run("RemoveRankedAttribute", func(t *testing.T) {
			newRankedAttributes := []models.AttributeName{"score"}
			updatedGame, err := models.NewGame(createdGame.GameID, createdGame.Description, createdGame.Attributes, newRankedAttributes)
			assert.NoError(t, err)

			result, err := gameService.UpdateGame(updatedGame)
			assert.NoError(t, err)
			assert.Equal(t, updatedGame, result)

			// Verify the game was updated and the attribute was removed
			retrievedGame, err := gameService.GetGame(createdGame.GameID)
			assert.NoError(t, err)
			assert.NotContains(t, retrievedGame.RankedAttributes, models.AttributeName("time"))
			assert.NotContains(t, retrievedGame.RankedAttributes, models.AttributeName("level"))

			// Verify that the leaderboard items for the removed attributes were deleted
			leaderboard, err := gameService.GetGameLeaderboard(createdGame.GameID, "time")
			assert.NoError(t, err)
			assert.Empty(t, leaderboard)

			leaderboard, err = gameService.GetGameLeaderboard(createdGame.GameID, "level")
			assert.NoError(t, err)
			assert.Empty(t, leaderboard)
		})

		// Clean up: Delete the game used for update tests
		err = gameService.DeleteGame(createdGame.GameID)
		assert.NoError(t, err)
	})

	// Test DeleteGame
	t.Run("DeleteGame", func(t *testing.T) {
		newGame, err := models.NewGame("tempgame", "Temporary game", []models.AttributeName{"score"}, []models.AttributeName{"score"})
		assert.NoError(t, err)
		createdGame, err := gameService.CreateGame(newGame)
		assert.NoError(t, err)

		// Add some items to the leaderboard for tempgame
		err = leaderboardRepo.AddLeaderboardItem(models.GameID("tempgame"), models.UserID("user1"), models.AttributeName("score"), 100, nil)
		assert.NoError(t, err)
		err = leaderboardRepo.AddLeaderboardItem(models.GameID("tempgame"), models.UserID("user2"), models.AttributeName("score"), 200, nil)
		assert.NoError(t, err)

		// Verify leaderboard items exist
		leaderboard, err := gameService.GetGameLeaderboard(models.GameID("tempgame"), models.AttributeName("score"))
		assert.NoError(t, err)
		assert.Equal(t, 2, len(leaderboard))

		err = gameService.DeleteGame(createdGame.GameID)
		assert.NoError(t, err)

		// Verify game is deleted
		_, err = gameService.GetGame(createdGame.GameID)
		assert.Error(t, err)

		// Verify leaderboard items are also deleted
		leaderboard, err = gameService.GetGameLeaderboard(models.GameID("tempgame"), models.AttributeName("score"))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(leaderboard))
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
