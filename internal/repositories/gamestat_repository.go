package repositories

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mquan1409/game-api/internal/models"
)

type GameStatRepository interface {
	GetGameStat(userID models.UserID, gameID models.GameID) (*models.GameStat, error)
	CreateGameStat(gameStat *models.GameStat, tx *dynamodb.TransactWriteItemsInput) error
	UpdateGameStat(gameStat *models.GameStat, tx *dynamodb.TransactWriteItemsInput) error
	DeleteGameStat(userID models.UserID, gameID models.GameID, tx *dynamodb.TransactWriteItemsInput) error
}
