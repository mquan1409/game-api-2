package services

import (
	"github.com/mquan1409/game-api/internal/models"
	"github.com/mquan1409/game-api/internal/repositories"
)

type MatchServiceImpl struct {
	matchRepository      repositories.MatchRepository
	gameRepository       repositories.GameRepository
	gameStatRepository   repositories.GameStatRepository
	leaderboardRepository repositories.LeaderboardRepository
}

func NewMatchServiceImpl(
	matchRepository repositories.MatchRepository,
	gameRepository repositories.GameRepository,
	gameStatRepository repositories.GameStatRepository,
	leaderboardRepository repositories.LeaderboardRepository,
) MatchService {
	return &MatchServiceImpl{
		matchRepository:      matchRepository,
		gameRepository:       gameRepository,
		gameStatRepository:   gameStatRepository,
		leaderboardRepository: leaderboardRepository,
	}
}

func (s *MatchServiceImpl) GetMatch(gameID models.GameID, matchID models.MatchID, dateID models.DateID) (*models.Match, error) {
	return s.matchRepository.GetMatch(gameID, matchID, dateID)
}

func (s *MatchServiceImpl) GetMatchesByGameAndDate(gameID models.GameID, dateID models.DateID) ([]*models.Match, error) {
	return s.matchRepository.GetMatchesByGameAndDate(gameID, dateID)
}

func (s *MatchServiceImpl) CreateMatch(match *models.Match) (*models.Match, error) {
	game, err := s.gameRepository.GetGame(match.GameID)
	if err != nil {
		return nil, err
	}

	createdMatch, err := s.matchRepository.CreateMatch(match, nil)
	if err != nil {
		return nil, err
	}

	// Update GameStats and Leaderboards for each player
	for userID, attributes := range match.PlayerAttributesMap {
		gameStat, err := s.gameStatRepository.GetGameStat(userID, match.GameID)
		if err != nil {
			// If GameStat doesn't exist, create a new one with all attributes initialized to 0
			initialAttributes := models.AttributesStatsMap{}
			for _, attr := range game.Attributes {
				initialAttributes[attr] = 0
			}
			gameStat, err = models.NewGameStat(userID, match.GameID, initialAttributes)
			if err != nil {
				return nil, err
			}
		}

		// Update Leaderboards only for ranked attributes
		for _, attr := range game.RankedAttributes {
			if value, exists := attributes[attr]; exists {
				oldSum := gameStat.GameAttributes[attr]
				newSum := oldSum + value
				if err := s.leaderboardRepository.UpdateLeaderboardItem(match.GameID, userID, attr, newSum, oldSum, nil); err != nil {
					return nil, err
				}
			}
		}

		// Update GameStat
		for attrName, value := range attributes {
			if currentValue, exists := gameStat.GameAttributes[attrName]; exists {
				gameStat.GameAttributes[attrName] = currentValue + value
			} else {
				gameStat.GameAttributes[attrName] = value
			}
		}

		if err := s.gameStatRepository.UpdateGameStat(gameStat, nil); err != nil {
			return nil, err
		}

	}

	return createdMatch, nil
}

func (s *MatchServiceImpl) UpdateMatch(match *models.Match) (*models.Match, error) {
	oldMatch, err := s.matchRepository.GetMatch(match.GameID, match.MatchID, match.DateID)
	if err != nil {
		return nil, err
	}

	updatedMatch, err := s.matchRepository.UpdateMatch(match, nil)
	if err != nil {
		return nil, err
	}

	game, err := s.gameRepository.GetGame(match.GameID)
	if err != nil {
		return nil, err
	}

	// Update GameStats and Leaderboards for each player
	for userID, newAttributes := range match.PlayerAttributesMap {
		oldAttributes, _ := oldMatch.GetPlayerAttributes(userID)

		gameStat, err := s.gameStatRepository.GetGameStat(userID, match.GameID)
		if err != nil {
			return nil, err
		}
		// Update Leaderboards
		for _, attr := range game.RankedAttributes {
			if newValue, exists := newAttributes[attr]; exists {
				oldValue := oldAttributes[attr]
				oldSum := gameStat.GameAttributes[attr]
				newSum := oldSum + (newValue - oldValue)
				if err := s.leaderboardRepository.UpdateLeaderboardItem(match.GameID, userID, attr, newSum, oldSum, nil); err != nil {
					return nil, err
				}
			}
		}

		// Update GameStat
		for attrName, newValue := range newAttributes {
			oldValue := oldAttributes[attrName]
			oldSum := gameStat.GameAttributes[attrName]
			newSum := oldSum + (newValue - oldValue)
			gameStat.GameAttributes[attrName] = newSum
		}

		if err := s.gameStatRepository.UpdateGameStat(gameStat, nil); err != nil {
			return nil, err
		}

	}

	return updatedMatch, nil
}

func (s *MatchServiceImpl) DeleteMatch(gameID models.GameID, matchID models.MatchID, dateID models.DateID) error {
	match, err := s.matchRepository.GetMatch(gameID, matchID, dateID)
	if err != nil {
		return err
	}

	game, err := s.gameRepository.GetGame(gameID)
	if err != nil {
		return err
	}

	// Update GameStats and Leaderboards for each player
	for userID, attributes := range match.PlayerAttributesMap {
		gameStat, err := s.gameStatRepository.GetGameStat(userID, gameID)
		if err != nil {
			return err
		}
		// Update Leaderboards
		for _, attr := range game.RankedAttributes {
			if value, exists := attributes[attr]; exists {
				oldSum := gameStat.GameAttributes[attr]
				newSum := oldSum - value
				if err := s.leaderboardRepository.UpdateLeaderboardItem(gameID, userID, attr, newSum, oldSum, nil); err != nil {
					return err
				}
			}
		}

		// Update GameStat
		for attrName, value := range attributes {
			gameStat.GameAttributes[attrName] -= value
		}

		if err := s.gameStatRepository.UpdateGameStat(gameStat, nil); err != nil {
			return err
		}

	}

	return s.matchRepository.DeleteMatch(gameID, matchID, dateID, nil)
}

