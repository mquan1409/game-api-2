package repositories

import (
	"github.com/mquan1409/game-api/internal/models"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type UserRepository interface {
	GetUser(id models.UserID) (*models.User, error)
	GetUserBasicsByPrefix(prefix string) ([]*models.UserBasic, error)
	CreateUser(user *models.User, tx *dynamodb.TransactWriteItemsInput) (*models.User, error)
	UpdateUser(user *models.User, tx *dynamodb.TransactWriteItemsInput) (*models.User, error)
	DeleteUser(id *models.UserID, tx *dynamodb.TransactWriteItemsInput) error
}

