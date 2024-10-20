package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mquan1409/game-api/internal/models"
)

func SetupTestDB(cfg *models.Config) (*dynamodb.DynamoDB, error) {
	// Configure AWS session with endpoint and region
	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(cfg.DynamoDBEndpoint),
		Region:   aws.String(cfg.DynamoDBRegion),
	})
	if err != nil {
		return nil, err
	}

	// Create DynamoDB client using the configured session
	db := dynamodb.New(sess)

	return db, nil
}