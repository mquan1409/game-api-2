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
	"github.com/mquan1409/game-api/internal/services"
	"github.com/mquan1409/game-api/internal/utils"
)

func TestMatchAPI(t *testing.T) {
	// Load test configuration
	baseURL := config.LoadBaseURL()
	cfg := config.LoadConfig("development")
	// Setup
	db, err := utils.SetupTestDB(&cfg)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	matchRepo := repositories.NewDynamoDBMatchRepository(db, cfg.TableName)
	gameRepo := repositories.NewDynamoDBGameRepository(db, cfg.TableName)
	gameStatRepo := repositories.NewDynamoDBGameStatRepository(db, cfg.TableName)
	leaderboardRepo := repositories.NewDynamoDBLeaderboardRepository(db, cfg.TableName)
	matchService := services.NewMatchServiceImpl(matchRepo, gameRepo, gameStatRepo, leaderboardRepo)

	// Scan the entire table before tests
	beforeScan, err := utils.ScanEntireTable(db, cfg.TableName)
	if err != nil {
		t.Fatalf("Failed to scan table before tests: %v", err)
	}

	// Test GetMatch
	t.Run("GetMatch", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/matches/soccer/match1/2023-06-01", baseURL))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var match models.Match
		err = json.NewDecoder(resp.Body).Decode(&match)
		assert.NoError(t, err)
		assert.Equal(t, models.MatchID("match1"), match.MatchID)
		assert.Equal(t, models.DateID("2023-06-01"), match.DateID)
		assert.Equal(t, models.GameID("soccer"), match.GameID)
		assert.Equal(t, []string{"Team A", "Team B"}, match.TeamNames)
		assert.Equal(t, []int{2, 1}, match.TeamScores)
		assert.Equal(t, [][]string{{"user1", "user2"}, {"user3", "dianadancer"}}, match.TeamMembers)
	})

	// Test GetMatchesByGameAndDate
	t.Run("GetMatchesByGameAndDate", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/matches?game=soccer&date=2023-06-01", baseURL))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var matches []*models.Match
		err = json.NewDecoder(resp.Body).Decode(&matches)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(matches))
		assert.Equal(t, models.MatchID("match1"), matches[0].MatchID)
	})

	// Test CreateMatch
	t.Run("CreateMatch", func(t *testing.T) {
		newMatch, err := models.NewMatch(
			"testmatch",
			"2023-06-12",
			"soccer",
			[]string{"Team C", "Team D"},
			[]int{3, 2},
			[][]string{{"user1", "user2"}, {"user3", "dianadancer"}},
			map[models.UserID]models.AttributesStatsMap{
				"user1": {"goals": 2, "assists": 1},
				"user2": {"goals": 1, "assists": 2},
				"user3": {"goals": 1, "assists": 1},
				"dianadancer": {"goals": 1, "assists": 0},
			},
		)
		assert.NoError(t, err)
		jsonMatch, _ := json.Marshal(newMatch)

		resp, err := http.Post(fmt.Sprintf("%s/matches", baseURL), "application/json", bytes.NewBuffer(jsonMatch))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createdMatch models.Match
		err = json.NewDecoder(resp.Body).Decode(&createdMatch)
		assert.NoError(t, err)
		assert.Equal(t, newMatch.MatchID, createdMatch.MatchID)
		assert.Equal(t, newMatch.DateID, createdMatch.DateID)
		assert.Equal(t, newMatch.GameID, createdMatch.GameID)

		// Clean up: Delete the created match
		err = matchService.DeleteMatch("soccer", "testmatch", "2023-06-12")
		assert.NoError(t, err)
	})

	// Test UpdateMatch
	t.Run("UpdateMatch", func(t *testing.T) {
		// First, create a match to update
		newMatch, err := models.NewMatch(
			"updatematch",
			"2023-06-13",
			"soccer",
			[]string{"Team E", "Team F"},
			[]int{2, 2},
			[][]string{{"user1", "user2"}, {"user3", "dianadancer"}},
			map[models.UserID]models.AttributesStatsMap{
				"user1": {"goals": 1, "assists": 1},
				"user2": {"goals": 1, "assists": 1},
				"user3": {"goals": 1, "assists": 1},
				"dianadancer": {"goals": 1, "assists": 1},
			},
		)
		assert.NoError(t, err)
		_, err = matchService.CreateMatch(newMatch)
		assert.NoError(t, err)

		// Update the match
		updatedMatch, err := models.NewMatch(
			"updatematch",
			"2023-06-13",
			"soccer",
			[]string{"Team E", "Team F"},
			[]int{3, 2},
			[][]string{{"user1", "user2"}, {"user3", "dianadancer"}},
			map[models.UserID]models.AttributesStatsMap{
				"user1": {"goals": 2, "assists": 1},
				"user2": {"goals": 1, "assists": 2},
				"user3": {"goals": 1, "assists": 1},
				"dianadancer": {"goals": 1, "assists": 0},
			},
		)
		assert.NoError(t, err)
		jsonMatch, _ := json.Marshal(updatedMatch)

		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/matches/updatematch", baseURL), bytes.NewBuffer(jsonMatch))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var resultMatch models.Match
		json.NewDecoder(resp.Body).Decode(&resultMatch)
		assert.Equal(t, updatedMatch.TeamScores, resultMatch.TeamScores)
		assert.Equal(t, updatedMatch.PlayerAttributesMap, resultMatch.PlayerAttributesMap)

		// Clean up: Delete the updated match
		err = matchService.DeleteMatch("soccer", "updatematch", "2023-06-13")
		assert.NoError(t, err)
	})

	// Test DeleteMatch
	t.Run("DeleteMatch", func(t *testing.T) {
		// First, create a match to delete
		newMatch, err := models.NewMatch(
			"deletematch",
			"2023-06-14",
			"soccer",
			[]string{"Team G", "Team H"},
			[]int{1, 1},
			[][]string{{"user1", "user2"}, {"user3", "dianadancer"}},
			map[models.UserID]models.AttributesStatsMap{
				"user1": {"goals": 1, "assists": 0},
				"user2": {"goals": 0, "assists": 1},
				"user3": {"goals": 1, "assists": 0},
				"dianadancer": {"goals": 0, "assists": 1},
			},
		)
		assert.NoError(t, err)
		_, err = matchService.CreateMatch(newMatch)
		assert.NoError(t, err)

		// Delete the match
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/matches/soccer/deletematch/2023-06-14", baseURL), nil)
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Verify match is deleted
		resp, err = http.Get(fmt.Sprintf("%s/matches/soccer/deletematch/2023-06-14", baseURL))
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

