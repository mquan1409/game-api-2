package services

import (
	"github.com/mquan1409/game-api/internal/models"
)

type GameService interface {
	GetGame(id models.GameID) (*models.Game, error)
	GetGameLeaderboard(gameID models.GameID, attribute models.AttributeName) (*models.LeaderBoard, error)
	GetBoundedGameLeaderboard(gameID models.GameID, attribute models.AttributeName, limit int) (*models.BoundedLeaderboard, error)
	CreateGame(game *models.Game) (*models.Game, error)
	UpdateGame(game *models.Game) (*models.Game, error)
	DeleteGame(id models.GameID) error
}