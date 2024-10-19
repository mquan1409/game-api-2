package repositories

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mquan1409/game-api/internal/models"
)

type LeaderboardRepository interface {
	GetLeaderboard(gameID models.GameID, attr models.AttributeName) (models.LeaderBoard, error)
	GetBoundedLeaderboard(gameID models.GameID, attr models.AttributeName, limit int) (models.BoundedLeaderboard, error)
	AddLeaderboardItem(gameID models.GameID, userID models.UserID, attr models.AttributeName, value models.AttributeStat, tx *dynamodb.TransactWriteItemsInput) error
	UpdateLeaderboardItem(gameID models.GameID, userID models.UserID, attr models.AttributeName, value models.AttributeStat, oldValue models.AttributeStat, tx *dynamodb.TransactWriteItemsInput) error
	DeleteLeaderboardItem(gameID models.GameID, userID models.UserID, attr models.AttributeName, oldValue models.AttributeStat, tx *dynamodb.TransactWriteItemsInput) error
	DeleteLeaderboardItemsByGame(gameID models.GameID, tx *dynamodb.TransactWriteItemsInput) error
	DeleteLeaderboardItemsByGameAndUser(gameID models.GameID, userID models.UserID, tx *dynamodb.TransactWriteItemsInput) error
	DeleteLeaderboardItemsByGameAndAttribute(gameID models.GameID, attr models.AttributeName, tx *dynamodb.TransactWriteItemsInput) error
}
