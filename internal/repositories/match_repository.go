package repositories

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mquan1409/game-api/internal/models"
)

type MatchRepository interface {
	GetMatch(gameID models.GameID, matchID models.MatchID, dateID models.DateID) (*models.Match, error)
	GetMatchesByGameAndDate(gameID models.GameID, dateID models.DateID) ([]*models.Match, error)
	CreateMatch(match *models.Match, tx *dynamodb.TransactWriteItemsInput) (*models.Match, error)
	UpdateMatch(match *models.Match, tx *dynamodb.TransactWriteItemsInput) (*models.Match, error)
	DeleteMatch(gameID models.GameID, matchID models.MatchID, dateID models.DateID, tx *dynamodb.TransactWriteItemsInput) error
}
