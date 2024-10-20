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
	"github.com/mquan1409/game-api/internal/repositories"
	"github.com/mquan1409/game-api/internal/utils"
)

func TestGameAPI(t *testing.T) {
	// Load test configuration
	baseURL := config.LoadBaseURL()
	cfg := config.LoadConfig("development")
	// Setup
	db, err := utils.SetupTestDB(&cfg)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	gameRepo := repositories.NewDynamoDBGameRepository(db, cfg.TableName)
	leaderboardRepo := repositories.NewDynamoDBLeaderboardRepository(db, cfg.TableName)

	// Scan the entire table before tests
	beforeScan, err := utils.ScanEntireTable(db, cfg.TableName)
	if err != nil {
		t.Fatalf("Failed to scan table before tests: %v", err)
	}

	// Test GetGame
	t.Run("GetGame", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/games/soccer", baseURL))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var game models.Game
		err = json.NewDecoder(resp.Body).Decode(&game)
		assert.NoError(t, err)
		assert.Equal(t, models.GameID("soccer"), game.GameID)
		assert.Equal(t, "Soccer", game.Description)
		assert.Contains(t, game.Attributes, models.AttributeName("elo"))
		assert.Contains(t, game.Attributes, models.AttributeName("goals"))
		assert.Contains(t, game.RankedAttributes, models.AttributeName("elo"))
	})

	// Test GetLeaderboard
	t.Run("GetLeaderboard", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/games/soccer/leaderboard/elo", baseURL))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var leaderboard models.LeaderBoard
		err = json.NewDecoder(resp.Body).Decode(&leaderboard)
		assert.NoError(t, err)
		assert.Equal(t, models.GameID("soccer"), leaderboard.GameID)
		assert.Equal(t, models.AttributeName("elo"), leaderboard.AttributeName)
		assert.Equal(t, 5, len(leaderboard.UserIDs))
		assert.Contains(t, leaderboard.UserIDs, models.UserID("user2"))
		assert.Contains(t, leaderboard.UserIDs, models.UserID("user1"))
		assert.Contains(t, leaderboard.UserIDs, models.UserID("user3"))
		assert.Contains(t, leaderboard.UserIDs, models.UserID("dianadancer"))
		assert.Contains(t, leaderboard.UserIDs, models.UserID("eveexplorer"))
	})

	// Test GetBoundedLeaderboard
	t.Run("GetBoundedLeaderboard", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/games/soccer/leaderboard/elo?limit=3", baseURL))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var leaderboard models.BoundedLeaderboard
		err = json.NewDecoder(resp.Body).Decode(&leaderboard)
		assert.NoError(t, err)
		assert.Equal(t, models.GameID("soccer"), leaderboard.GameID)
		assert.Equal(t, models.AttributeName("elo"), leaderboard.AttributeName)
		assert.Equal(t, 3, len(leaderboard.UserIDs))
		assert.Equal(t, 3, leaderboard.Limit)
	})

	// Test CreateGame
	t.Run("CreateGame", func(t *testing.T) {
		newGame, err := models.NewGame("testgame", "Test Game", []models.AttributeName{"score", "time"}, []models.AttributeName{"score"})
		assert.NoError(t, err)
		jsonGame, _ := json.Marshal(newGame)

		resp, err := http.Post(fmt.Sprintf("%s/games", baseURL), "application/json", bytes.NewBuffer(jsonGame))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createdGame models.Game
		err = json.NewDecoder(resp.Body).Decode(&createdGame)
		assert.NoError(t, err)
		assert.Equal(t, newGame.GameID, createdGame.GameID)
		assert.Equal(t, newGame.Description, createdGame.Description)

		// Clean up: Delete the created game
		err = gameRepo.DeleteGame(createdGame.GameID, nil)
		assert.NoError(t, err)
	})

	// Test UpdateGame
	t.Run("UpdateGame", func(t *testing.T) {
		// Update an existing game (soccer)
		updatedGame, err := models.NewGame("soccer", "Updated Soccer", []models.AttributeName{"elo", "goals", "assists", "shots_on_target", "passes_completed", "new_attribute"}, []models.AttributeName{"elo", "goals"})
		assert.NoError(t, err)
		jsonGame, _ := json.Marshal(updatedGame)

		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/games/soccer", baseURL), bytes.NewBuffer(jsonGame))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var resultGame models.Game
		json.NewDecoder(resp.Body).Decode(&resultGame)
		assert.Equal(t, updatedGame.Description, resultGame.Description)
		assert.Equal(t, updatedGame.Attributes, resultGame.Attributes)
		assert.Equal(t, updatedGame.RankedAttributes, resultGame.RankedAttributes)

		// Clean up: Revert the game to its original state
		originalGame, err := models.NewGame("soccer", "Soccer", []models.AttributeName{"elo", "goals", "assists", "shots_on_target", "passes_completed"}, []models.AttributeName{"elo"})
		assert.NoError(t, err)
		_, err = gameRepo.UpdateGame(originalGame, nil)
		assert.NoError(t, err)
	})

	// Test DeleteGame
	t.Run("DeleteGame", func(t *testing.T) {
		// First, create a game to delete
		newGame, err := models.NewGame("deletegame", "Game to Delete", []models.AttributeName{"score"}, []models.AttributeName{"score"})
		assert.NoError(t, err)
		_, err = gameRepo.CreateGame(newGame, nil)
		assert.NoError(t, err)

		// Delete the game
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/games/deletegame", baseURL), nil)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Verify game is deleted
		resp, err = http.Get(fmt.Sprintf("%s/games/deletegame", baseURL))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Clean up: Delete any remaining leaderboard items
		err = leaderboardRepo.DeleteLeaderboardItemsByGame(newGame.GameID, nil)
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

