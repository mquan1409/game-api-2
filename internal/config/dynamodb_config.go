package config

import (
	"os"
	"github.com/mquan1409/game-api/internal/models"
)

func LoadConfig(env string) models.Config {
	switch env {
	case "production":
		return loadProductionConfig()
	default:
		return loadDevelopmentConfig()
	}
}

func loadProductionConfig() models.Config {
	return models.Config{
		DynamoDBEndpoint: os.Getenv("DYNAMODB_ENDPOINT"),
		DynamoDBRegion:   os.Getenv("DYNAMODB_REGION"),
		TableName:        os.Getenv("DYNAMODB_TABLE"),
	}
}

func loadDevelopmentConfig() models.Config {
	return models.Config{
		DynamoDBEndpoint: os.Getenv("DYNAMODB_ENDPOINT"),
		DynamoDBRegion:   os.Getenv("DYNAMODB_REGION"),
		TableName:        os.Getenv("DYNAMODB_TABLE"),
	}
}
