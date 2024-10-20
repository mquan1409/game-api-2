package services

import (
	"github.com/mquan1409/game-api/internal/models"
	"github.com/mquan1409/game-api/internal/repositories"
	"github.com/mquan1409/game-api/internal/utils"
)

type GameServiceImpl struct {
	gameRepository        repositories.GameRepository
	leaderboardRepository repositories.LeaderboardRepository
}

func NewGameServiceImpl(gameRepository repositories.GameRepository, leaderboardRepository repositories.LeaderboardRepository) GameService {
	return &GameServiceImpl{
		gameRepository:        gameRepository,
		leaderboardRepository: leaderboardRepository,
	}
}

func (s *GameServiceImpl) GetGame(id models.GameID) (*models.Game, error) {
	return s.gameRepository.GetGame(id)
}

func (s *GameServiceImpl) GetGameLeaderboard(gameID models.GameID, attribute models.AttributeName) (*models.LeaderBoard, error) {
	leaderboard, err := s.leaderboardRepository.GetLeaderboard(gameID, attribute)
	if err != nil {
		return nil, err
	}
	return &leaderboard, nil
}

func (s *GameServiceImpl) GetBoundedGameLeaderboard(gameID models.GameID, attribute models.AttributeName, limit int) (*models.BoundedLeaderboard, error) {
	boundedLeaderboard, err := s.leaderboardRepository.GetBoundedLeaderboard(gameID, attribute, limit)
	if err != nil {
		return nil, err
	}
	return &boundedLeaderboard, nil
}

func (s *GameServiceImpl) CreateGame(game *models.Game) (*models.Game, error) {
	return s.gameRepository.CreateGame(game, nil)
}

func (s *GameServiceImpl) UpdateGame(game *models.Game) (*models.Game, error) {
	oldGame, err := s.gameRepository.GetGame(game.GameID)
	if err != nil {
		return nil, err
	}
	deletedAttributes := utils.Minus(oldGame.RankedAttributes, game.RankedAttributes)
	for _, attribute := range deletedAttributes {
		err := s.leaderboardRepository.DeleteLeaderboardItemsByGameAndAttribute(game.GameID, attribute, nil)
		if err != nil {
			return nil, err
		}
	}

	return s.gameRepository.UpdateGame(game, nil)
}

func (s *GameServiceImpl) DeleteGame(id models.GameID) error {
	err := s.leaderboardRepository.DeleteLeaderboardItemsByGame(id, nil)
	if err != nil {
		return err
	}
	return s.gameRepository.DeleteGame(id, nil)
}
