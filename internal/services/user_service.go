package services

import (
	"github.com/mquan1409/game-api/internal/models"
)

type UserService interface {
	GetUser(id models.UserID) (*models.User, error)
	GetUserBasicsByPrefix(prefix string) ([]*models.UserBasic, error)
	GetGameStat(userID models.UserID, gameID models.GameID) (*models.GameStat, error)	
	CreateUser(user *models.User) (*models.User, error)
	UpdateUser(user *models.User) (*models.User, error)
	DeleteUser(id *models.UserID) error
}



