package services

import (
	"github.com/mquan1409/game-api/internal/models"
	"github.com/mquan1409/game-api/internal/repositories"
)

type UserServiceImpl struct {
	userRepository repositories.UserRepository
	gamestatRepository repositories.GameStatRepository
}

func NewUserServiceImpl(userRepository repositories.UserRepository, gamestatRepository repositories.GameStatRepository) UserService {
	return &UserServiceImpl{
		userRepository: userRepository,
		gamestatRepository: gamestatRepository,
	}
}
func (s *UserServiceImpl) GetUser(id models.UserID) (*models.User, error) {
	return s.userRepository.GetUser(id)
}

func (s *UserServiceImpl) GetUserBasicsByPrefix(prefix string) ([]*models.UserBasic, error) {
	return s.userRepository.GetUserBasicsByPrefix(prefix)
}

func (s *UserServiceImpl) GetGameStat(userID models.UserID, gameID models.GameID) (*models.GameStat, error) {
	gameStat, err := s.gamestatRepository.GetGameStat(userID, gameID)
	if err != nil {
		return nil, err
	}
	return gameStat, nil
}

func (s *UserServiceImpl) CreateUser(user *models.User) (*models.User, error) {
	return s.userRepository.CreateUser(user, nil)
}

func (s *UserServiceImpl) UpdateUser(user *models.User) (*models.User, error) {
	return s.userRepository.UpdateUser(user, nil)
}

func (s *UserServiceImpl) DeleteUser(id *models.UserID) error {
	return s.userRepository.DeleteUser(id, nil)
}



