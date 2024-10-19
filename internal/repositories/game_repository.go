package repositories

import (
	"github.com/mquan1409/game-api/internal/models"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type GameRepository interface {
	GetGame(id models.GameID) (*models.Game, error)
	CreateGame(game *models.Game, tx *dynamodb.TransactWriteItemsInput) (*models.Game, error)
	UpdateGame(game *models.Game, tx *dynamodb.TransactWriteItemsInput) (*models.Game, error)
	DeleteGame(id models.GameID, tx *dynamodb.TransactWriteItemsInput) error
}

