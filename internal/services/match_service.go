package services

import (
	"github.com/mquan1409/game-api/internal/models"
)

type MatchService interface {
	GetMatch(gameID models.GameID, matchID models.MatchID, dateID models.DateID) (*models.Match, error)
	GetMatchesByGameAndDate(gameID models.GameID, dateID models.DateID) ([]*models.Match, error)
	CreateMatch(match *models.Match) (*models.Match, error)
	UpdateMatch(match *models.Match) (*models.Match, error)
	DeleteMatch(gameID models.GameID, matchID models.MatchID, dateID models.DateID) error
}

