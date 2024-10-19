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

func TestMatchService(t *testing.T) {
	// Load test configuration
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
		// Test with existing match data from add-matches.json
		retrievedMatch, err := matchService.GetMatch("soccer", "match1", "2023-06-01")
		assert.NoError(t, err)
		assert.NotNil(t, retrievedMatch)

		// Verify match details
		assert.Equal(t, models.MatchID("match1"), retrievedMatch.MatchID)
		assert.Equal(t, models.DateID("2023-06-01"), retrievedMatch.DateID)
		assert.Equal(t, models.GameID("soccer"), retrievedMatch.GameID)
		assert.Equal(t, []string{"Team A", "Team B"}, retrievedMatch.TeamNames)
		assert.Equal(t, []int{2, 1}, retrievedMatch.TeamScores)
		assert.Equal(t, [][]string{{"user1", "user2"}, {"user3", "dianadancer"}}, retrievedMatch.TeamMembers)

		// Verify player attributes
		expectedAttributes := map[models.UserID]models.AttributesStatsMap{
			"user1":       {"goals": 1, "assists": 1, "shots_on_target": 3, "passes_completed": 20},
			"user2":       {"goals": 1, "assists": 0, "shots_on_target": 2, "passes_completed": 15},
			"user3":       {"goals": 1, "assists": 0, "shots_on_target": 2, "passes_completed": 18},
			"dianadancer": {"goals": 0, "assists": 1, "shots_on_target": 1, "passes_completed": 22},
		}
		assert.Equal(t, expectedAttributes, retrievedMatch.PlayerAttributesMap)
	})

	// Test GetMatchesByGameAndDate
	t.Run("GetMatchesByGameAndDate", func(t *testing.T) {
		// Create two test matches
		match1, _ := models.NewMatch("testmatch2", "2023-06-11", "soccer", []string{"Team C", "Team D"}, []int{3, 3}, [][]string{{"user1", "user3"}, {"user2", "dianadancer"}}, nil)
		match2, _ := models.NewMatch("testmatch3", "2023-06-11", "soccer", []string{"Team E", "Team F"}, []int{1, 0}, [][]string{{"user1", "dianadancer"}, {"user2", "user3"}}, nil)

		_, err = matchService.CreateMatch(match1)
		assert.NoError(t, err)
		_, err = matchService.CreateMatch(match2)
		assert.NoError(t, err)

		matches, err := matchService.GetMatchesByGameAndDate("soccer", "2023-06-11")
		assert.NoError(t, err)
		assert.NotEmpty(t, matches)
		assert.Equal(t, 2, len(matches))
		assert.Contains(t, []models.MatchID{"testmatch2", "testmatch3"}, matches[0].MatchID)
		assert.Contains(t, []models.MatchID{"testmatch2", "testmatch3"}, matches[1].MatchID)

		// Clean up: Delete the created matches
		err = matchService.DeleteMatch("soccer", "testmatch2", "2023-06-11")
		assert.NoError(t, err)
		err = matchService.DeleteMatch("soccer", "testmatch3", "2023-06-11")
		assert.NoError(t, err)
	})
	// Test CreateMatch
	t.Run("CreateMatch", func(t *testing.T) {
		// Create a new match
		newMatch, _ := models.NewMatch("testmatch1", "2023-06-10", "soccer", []string{"Team X", "Team Y"}, []int{2, 1}, [][]string{{"user1", "user2"}, {"user3", "dianadancer"}}, map[models.UserID]models.AttributesStatsMap{
			"user1":       {"goals": 1, "assists": 1, "shots_on_target": 2, "passes_completed": 15, "elo": 10},
			"user2":       {"goals": 1, "assists": 0, "shots_on_target": 1, "passes_completed": 10, "elo": 8},
			"user3":       {"goals": 1, "assists": 0, "shots_on_target": 2, "passes_completed": 12, "elo": 5},
			"dianadancer": {"goals": 0, "assists": 1, "shots_on_target": 1, "passes_completed": 18, "elo": 3},
		})

		// Create the match
		createdMatch, err := matchService.CreateMatch(newMatch)
		if err != nil {
			t.Fatalf("Failed to create match: %v", err)
		}
		assert.NotNil(t, createdMatch)

		// Verify the match was created correctly
		retrievedMatch, err := matchService.GetMatch("soccer", "testmatch1", "2023-06-10")
		if err != nil {
			t.Fatalf("Failed to retrieve created match: %v", err)
		}
		assert.Equal(t, newMatch, retrievedMatch)

		// Verify GameStats were updated correctly
		expectedGameStats := map[models.UserID]models.AttributesStatsMap{
			"user1": {
				"elo":              12,
				"goals":            2,
				"assists":          2,
				"shots_on_target":  5,
				"passes_completed": 35,
			},
			"user2": {
				"elo":              10,
				"goals":            2,
				"assists":          0,
				"shots_on_target":  3,
				"passes_completed": 25,
			},
			"user3": {
				"elo":              6,
				"goals":            2,
				"assists":          0,
				"shots_on_target":  4,
				"passes_completed": 30,
			},
			"dianadancer": {
				"elo":              4,
				"goals":            0,
				"assists":          2,
				"shots_on_target":  2,
				"passes_completed": 40,
			},
		}

		for userID, expectedAttributes := range expectedGameStats {
			gameStat, err := gameStatRepo.GetGameStat(userID, "soccer")
			if err != nil {
				t.Fatalf("Failed to get game stat for user %s: %v", userID, err)
			}
			assert.Equal(t, expectedAttributes, gameStat.GameAttributes)
		}

		// Verify Leaderboard was updated only for 'elo'
		leaderboard, err := leaderboardRepo.GetLeaderboard("soccer", "elo")
		assert.NoError(t, err)
		assert.NotEmpty(t, leaderboard.UserIDs)

		assert.Equal(t, 5, len(leaderboard.UserIDs)) // 5 users in the leaderboard as per add-leaderboard.json

		// Verify the order of users in the leaderboard
		expectedOrder := []models.UserID{"user1", "user2", "user3", "dianadancer", "eveexplorer"}
		assert.Equal(t, expectedOrder, leaderboard.UserIDs)

		// Clean up: Delete the created match and reset GameStats and Leaderboard
		err = matchRepo.DeleteMatch(models.GameID("soccer"), models.MatchID("testmatch1"), models.DateID("2023-06-10"), nil)
		if err != nil {
			t.Fatalf("Failed to delete match: %v", err)
		}

		for userID, attributes := range newMatch.PlayerAttributesMap {
			gameStat, err := gameStatRepo.GetGameStat(userID, "soccer")
			assert.NoError(t, err)
			for attrName, value := range attributes {
				gameStat.GameAttributes[attrName] -= value
			}
			err = gameStatRepo.UpdateGameStat(gameStat, nil)
			assert.NoError(t, err)

			// Update 'elo' in leaderboard back to its old value
			oldEloValue := map[models.UserID]models.AttributeStat{
				"user1":       2,
				"user2":       2,
				"user3":       1,
				"dianadancer": 1,
			}[userID]
			err = leaderboardRepo.UpdateLeaderboardItem("soccer", userID, "elo", oldEloValue, expectedGameStats[userID]["elo"], nil)
			assert.NoError(t, err)
		}

		// Verify the leaderboard is back to its original state
		leaderboard, err = leaderboardRepo.GetLeaderboard("soccer", "elo")
		assert.NoError(t, err)
		assert.Equal(t, 5, len(leaderboard.UserIDs))
		assert.Equal(t, []models.UserID{"user2", "user1", "user3", "dianadancer", "eveexplorer"}, leaderboard.UserIDs)

		// Reset GameStats to original state
		originalGameStats := map[models.UserID]models.AttributesStatsMap{
			"user1": {
				"elo":              2,
				"goals":            1,
				"assists":          1,
				"shots_on_target":  3,
				"passes_completed": 20,
			},
			"user2": {
				"elo":              2,
				"goals":            1,
				"assists":          0,
				"shots_on_target":  2,
				"passes_completed": 15,
			},
			"user3": {
				"elo":              1,
				"goals":            1,
				"assists":          0,
				"shots_on_target":  2,
				"passes_completed": 18,
			},
			"dianadancer": {
				"elo":              1,
				"goals":            0,
				"assists":          1,
				"shots_on_target":  1,
				"passes_completed": 22,
			},
		}

		for userID, originalAttributes := range originalGameStats {
			gameStat, err := gameStatRepo.GetGameStat(userID, "soccer")
			if err != nil {
				t.Fatalf("Failed to get game stat for user %s: %v", userID, err)
			}
			gameStat.GameAttributes = originalAttributes
			err = gameStatRepo.UpdateGameStat(gameStat, nil)
			if err != nil {
				t.Fatalf("Failed to update game stat for user %s: %v", userID, err)
			}
		}

		// Verify GameStats are back to their original state
		for userID, expectedAttributes := range originalGameStats {
			gameStat, err := gameStatRepo.GetGameStat(userID, "soccer")
			assert.NoError(t, err)
			assert.Equal(t, expectedAttributes, gameStat.GameAttributes)
		}
	})
	t.Run("DeleteMatch", func(t *testing.T) {
		newMatch, _ := models.NewMatch("testmatch5", "2023-06-13", "soccer", []string{"Team I", "Team J"}, []int{1, 1}, [][]string{{"user1", "user2"}, {"user3", "dianadancer"}}, nil)
		_, err := matchService.CreateMatch(newMatch)
		assert.NoError(t, err)

		// Delete the match
		err = matchService.DeleteMatch("soccer", "testmatch5", "2023-06-13")
		assert.NoError(t, err)

		// Verify the match was deleted
		_, err = matchService.GetMatch("soccer", "testmatch5", "2023-06-13")
		assert.Error(t, err)
	})

	// Test UpdateMatch
	t.Run("UpdateAndDeleteMatch", func(t *testing.T) {
		newMatch, _ := models.NewMatch("testmatch4", "2023-06-12", "soccer", []string{"Team G", "Team H"}, []int{2, 2}, [][]string{{"user1", "user2"}, {"user3", "dianadancer"}}, map[models.UserID]models.AttributesStatsMap{
			"user1": {"elo": 2, "goals": 1, "assists": 1},
			"user2": {"elo": 1, "goals": 1, "assists": 0},
			"user3": {"elo": 0, "goals": 0, "assists": 1},
			"dianadancer": {"elo": 0, "goals": 0, "assists": 0},
		})
		createdMatch, err := matchService.CreateMatch(newMatch)
		assert.NoError(t, err)

		// Update match scores and player attributes
		createdMatch.TeamScores = []int{3, 2}
		createdMatch.PlayerAttributesMap = map[models.UserID]models.AttributesStatsMap{
			"user1": {"elo": 2, "goals": 2, "assists": 1},
			"user2": {"elo": 1, "goals": 1, "assists": 1},
			"user3": {"elo": 1, "goals": 1, "assists": 1},
			"dianadancer": {"elo": 1, "goals": 1, "assists": 0},
		}
		updatedMatch, err := matchService.UpdateMatch(createdMatch)
		assert.NoError(t, err)
		assert.Equal(t, []int{3, 2}, updatedMatch.TeamScores)

		// Verify the match was updated
		retrievedMatch, err := matchService.GetMatch("soccer", "testmatch4", "2023-06-12")
		assert.NoError(t, err)
		assert.Equal(t, []int{3, 2}, retrievedMatch.TeamScores)

		// Verify GameStats were updated correctly
		expectedGameStats := map[models.UserID]models.AttributesStatsMap{
			"user1": {"elo": 4, "goals": 3, "assists": 2},
			"user2": {"elo": 3, "goals": 2, "assists": 1},
			"user3": {"elo": 2, "goals": 2, "assists": 1},
			"dianadancer": {"elo": 2, "goals": 1, "assists": 1},
		}
		for userID, expectedStats := range expectedGameStats {
			gameStat, err := gameStatRepo.GetGameStat(userID, "soccer")
			assert.NoError(t, err)
			for attr, value := range expectedStats {
				assert.Equal(t, value, gameStat.GameAttributes[attr])
				if t.Failed() {
					t.Fatalf("GameStat for user %s attribute %s is not updated correctly. Expected %d, got %d", userID, attr, value, gameStat.GameAttributes[attr])
				}
			}
		}

		// Verify Leaderboard was updated correctly
		leaderboard, err := leaderboardRepo.GetLeaderboard("soccer", "elo")
		assert.NoError(t, err)
		assert.Equal(t, models.GameID("soccer"), leaderboard.GameID)
		assert.Equal(t, models.AttributeName("elo"), leaderboard.AttributeName)
		assert.Equal(t, 5, len(leaderboard.UserIDs))
		assert.Equal(t, models.UserID("user1"), leaderboard.UserIDs[0])
		assert.Equal(t, models.UserID("user2"), leaderboard.UserIDs[1])
		assert.Contains(t, []models.UserID{"user3", "dianadancer"}, leaderboard.UserIDs[2])
		assert.Contains(t, []models.UserID{"user3", "dianadancer"}, leaderboard.UserIDs[3])

		// Clean up: Delete the created match
		err = matchService.DeleteMatch("soccer", "testmatch4", "2023-06-12")
		assert.NoError(t, err)
	})
	// Scan the entire table after tests
	afterScan, err := utils.ScanEntireTable(db, cfg.TableName)
	if err != nil {
		t.Fatalf("Failed to scan table after tests: %v", err)
	}

	// Compare before and after scans
	if !t.Failed() {
		assert.Equal(t, beforeScan, afterScan, "MatchService Test: The database state has changed after running tests")
	}

	// Test DeleteMatch

}
