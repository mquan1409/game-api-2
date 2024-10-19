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

func TestMatchRepository(t *testing.T) {
	// Load test configuration
	cfg := config.LoadConfig("development")

	// Setup
	db, err := utils.SetupTestDB(&cfg)
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	repo := repositories.NewDynamoDBMatchRepository(db, cfg.TableName)

	// Scan the entire table before tests
	beforeScan, err := utils.ScanEntireTable(db, cfg.TableName)
	if err != nil {
		t.Fatalf("Failed to scan table before tests: %v", err)
	}

	// Test GetMatch
	t.Run("GetMatch", func(t *testing.T) {
		match, err := repo.GetMatch("soccer", "match1", "2023-06-01")
		assert.NoError(t, err)
		assert.NotNil(t, match)
		assert.Equal(t, models.MatchID("match1"), match.MatchID)
		assert.Equal(t, models.DateID("2023-06-01"), match.DateID)
		assert.Equal(t, models.GameID("soccer"), match.GameID)
		assert.Equal(t, []string{"Team A", "Team B"}, match.TeamNames)
		assert.Equal(t, []int{2, 1}, match.TeamScores)
	})

	// Test CreateMatch
	t.Run("CreateMatch", func(t *testing.T) {
		newMatch, err := models.NewMatch(models.MatchID("newmatch"), models.DateID("2023-06-04"), models.GameID("basketball"), []string{"Team C", "Team D"}, []int{80, 75}, [][]string{{"user1", "user2"}, {"user3", "dianadancer"}}, make(map[models.UserID]models.AttributesStatsMap))
		assert.NoError(t, err)
		createdMatch, err := repo.CreateMatch(newMatch, nil)
		assert.NoError(t, err)
		assert.Equal(t, newMatch, createdMatch)

		// Verify the match was actually created
		fetchedMatch, err := repo.GetMatch("basketball", "newmatch", "2023-06-04")
		assert.NoError(t, err)
		assert.Equal(t, newMatch, fetchedMatch)

		// Clean up: delete the created match
		err = repo.DeleteMatch("basketball", "newmatch", "2023-06-04", nil)
		assert.NoError(t, err)
	})

	// Test UpdateMatch
	t.Run("UpdateMatch", func(t *testing.T) {
		match, err := repo.GetMatch("pool", "match2", "2023-06-02")
		assert.NoError(t, err)

		originalScores := match.TeamScores

		// Update the match
		match.TeamScores = []int{4, 2}
		updatedMatch, err := repo.UpdateMatch(match, nil)
		assert.NoError(t, err)
		assert.Equal(t, match, updatedMatch)

		// Verify the match was actually updated
		fetchedMatch, err := repo.GetMatch("pool", "match2", "2023-06-02")
		assert.NoError(t, err)
		assert.Equal(t, []int{4, 2}, fetchedMatch.TeamScores)

		// Clean up: revert the match to original state
		match.TeamScores = originalScores
		_, err = repo.UpdateMatch(match, nil)
		assert.NoError(t, err)
	})

	// Test DeleteMatch
	t.Run("DeleteMatch", func(t *testing.T) {
		// Create a temporary match to delete
		tempMatch, err := models.NewMatch(models.MatchID("tempmatch"), models.DateID("2023-06-06"), models.GameID("pickleball"), []string{"Team G", "Team H"}, []int{4, 4}, [][]string{{"user1", "user2"}, {"user3", "dianadancer"}}, make(map[models.UserID]models.AttributesStatsMap))
		assert.NoError(t, err)
		_, err = repo.CreateMatch(tempMatch, nil)
		assert.NoError(t, err)

		err = repo.DeleteMatch("pickleball", "tempmatch", "2023-06-06", nil)
		assert.NoError(t, err)

		// Verify match is deleted
		_, err = repo.GetMatch("pickleball", "tempmatch", "2023-06-06")
		assert.Error(t, err)
	})

	// Test GetMatchesByGameIDAndDateID
	t.Run("GetMatchesByGameAndDate", func(t *testing.T) {
		// Create two new matches with the same date
		match1, err := models.NewMatch(models.MatchID("testmatch1"), models.DateID("2023-06-07"), models.GameID("tennis"), []string{"Player A", "Player B"}, []int{6, 4}, [][]string{{"user1"}, {"user2"}}, make(map[models.UserID]models.AttributesStatsMap))
		assert.NoError(t, err)
		_, err = repo.CreateMatch(match1, nil)
		assert.NoError(t, err)

		match2, err := models.NewMatch(models.MatchID("testmatch2"), models.DateID("2023-06-07"), models.GameID("tennis"), []string{"Player C", "Player D"}, []int{7, 5}, [][]string{{"user3"}, {"user4"}}, make(map[models.UserID]models.AttributesStatsMap))
		assert.NoError(t, err)
		_, err = repo.CreateMatch(match2, nil)
		assert.NoError(t, err)

		// Test GetMatchesByGameIDAndDateID
		matches, err := repo.GetMatchesByGameAndDate("tennis", "2023-06-07")
		assert.NoError(t, err)
		assert.Len(t, matches, 2)
		assert.Contains(t, matches, match1)
		assert.Contains(t, matches, match2)

		// Clean up: delete the test matches
		err = repo.DeleteMatch("tennis", "testmatch1", "2023-06-07", nil)
		assert.NoError(t, err)
		err = repo.DeleteMatch("tennis", "testmatch2", "2023-06-07", nil)
		assert.NoError(t, err)

		// Verify matches are deleted
		matches, err = repo.GetMatchesByGameAndDate("tennis", "2023-06-07")
		assert.NoError(t, err)
		assert.Len(t, matches, 0)
	})

	// Test CreateMatch and UpdateMatch in the same transaction
	t.Run("CreateAndUpdateInTransaction", func(t *testing.T) {
		tx := &dynamodb.TransactWriteItemsInput{}

		// Create a new match
		newMatch, err := models.NewMatch(models.MatchID("newmatch2"), models.DateID("2023-06-05"), models.GameID("soccer"), []string{"Team E", "Team F"}, []int{3, 3}, [][]string{{"user1", "user2"}, {"user3", "eveexplorer"}}, make(map[models.UserID]models.AttributesStatsMap))
		assert.NoError(t, err)
		createdMatch, err := repo.CreateMatch(newMatch, tx)
		assert.NoError(t, err)
		assert.Equal(t, newMatch, createdMatch)

		// Update an existing match
		existingMatch, err := repo.GetMatch("soccer", "match1", "2023-06-01")
		assert.NoError(t, err)
		originalScores := existingMatch.TeamScores
		existingMatch.TeamScores = []int{3, 1}
		updatedMatch, err := repo.UpdateMatch(existingMatch, tx)
		assert.NoError(t, err)
		assert.Equal(t, existingMatch, updatedMatch)

		// Execute the transaction
		_, err = db.TransactWriteItems(tx)
		assert.NoError(t, err)

		// Verify the new match was created
		createdMatch, err = repo.GetMatch("soccer", "newmatch2", "2023-06-05")
		assert.NoError(t, err)
		assert.Equal(t, newMatch, createdMatch)

		// Verify the existing match was updated
		updatedMatch, err = repo.GetMatch("soccer", "match1", "2023-06-01")
		assert.NoError(t, err)
		assert.Equal(t, []int{3, 1}, updatedMatch.TeamScores)

		// Clean up: delete the newly created match
		err = repo.DeleteMatch("soccer", "newmatch2", "2023-06-05", nil)
		assert.NoError(t, err)

		// Revert the updated match
		existingMatch.TeamScores = originalScores
		_, err = repo.UpdateMatch(existingMatch, nil)
		assert.NoError(t, err)
	})

	// Test DeleteMatch and UpdateMatch in the same transaction
	t.Run("DeleteAndUpdateInTransaction", func(t *testing.T) {
		tx := &dynamodb.TransactWriteItemsInput{}

		// Create a temporary match to delete
		tempMatch, err := models.NewMatch(models.MatchID("tempmatch"), models.DateID("2023-06-06"), models.GameID("pool"), []string{"Team G", "Team H"}, []int{4, 4}, [][]string{{"user1", "user2"}, {"user3", "dianadancer"}}, make(map[models.UserID]models.AttributesStatsMap))
		assert.NoError(t, err)
		_, err = repo.CreateMatch(tempMatch, nil)
		assert.NoError(t, err)

		// Delete the temporary match
		err = repo.DeleteMatch("pool", "tempmatch", "2023-06-06", tx)
		assert.NoError(t, err)

		// Update an existing match
		existingMatch, err := repo.GetMatch("pool", "match2", "2023-06-02")
		assert.NoError(t, err)
		originalScores := existingMatch.TeamScores
		existingMatch.TeamScores = []int{5, 2}
		updatedMatch, err := repo.UpdateMatch(existingMatch, tx)
		assert.NoError(t, err)
		assert.Equal(t, existingMatch, updatedMatch)

		// Execute the transaction
		_, err = db.TransactWriteItems(tx)
		assert.NoError(t, err)

		// Verify the temporary match was deleted
		_, err = repo.GetMatch("pool", "tempmatch", "2023-06-06")
		assert.Error(t, err)

		// Verify the existing match was updated
		updatedMatch, err = repo.GetMatch("pool", "match2", "2023-06-02")
		assert.NoError(t, err)
		assert.Equal(t, []int{5, 2}, updatedMatch.TeamScores)

		// Revert the updated match
		existingMatch.TeamScores = originalScores
		_, err = repo.UpdateMatch(existingMatch, nil)
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
