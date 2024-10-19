package config

import "os"

type Config struct {
	DynamoDBEndpoint string
	DynamoDBRegion   string
	TableName        string
	// Add other configuration fields as needed
}

func LoadConfig(env string) Config {
	switch env {
	case "production":
		return loadProductionConfig()
	default:
		return loadDevelopmentConfig()
	}
}

func loadProductionConfig() Config {
	return Config{
		DynamoDBEndpoint: os.Getenv("DYNAMODB_ENDPOINT"),
		DynamoDBRegion:   os.Getenv("DYNAMODB_REGION"),
		TableName:        os.Getenv("DYNAMODB_TABLE"),
	}
}

func loadDevelopmentConfig() Config {
	return Config{
		DynamoDBEndpoint: os.Getenv("DYNAMODB_ENDPOINT"),
		DynamoDBRegion:   os.Getenv("DYNAMODB_REGION"),
		TableName:        os.Getenv("DYNAMODB_TABLE"),
	}
}
