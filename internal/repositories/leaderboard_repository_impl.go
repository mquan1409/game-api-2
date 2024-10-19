package repositories

import (
	"errors"
	"fmt"


	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/mquan1409/game-api/internal/models"
	"github.com/mquan1409/game-api/internal/utils"
)

type DynamoDBLeaderboardRepository struct {
	db        *dynamodb.DynamoDB
	tableName string
}

func NewDynamoDBLeaderboardRepository(db *dynamodb.DynamoDB, tableName string) LeaderboardRepository {
	return &DynamoDBLeaderboardRepository{db: db, tableName: tableName}
}

func (r *DynamoDBLeaderboardRepository) GetLeaderboard(gameID models.GameID, attr models.AttributeName) (models.LeaderBoard, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("Id = :id AND begins_with(#range, :attr)"),
		ExpressionAttributeNames: map[string]*string{
			"#range": aws.String("Range"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id":   {S: aws.String(fmt.Sprintf("Leaderboard.%s", gameID))},
			":attr": {S: aws.String(fmt.Sprintf("%s.", attr))},
		},
		ScanIndexForward: aws.Bool(false),
	}

	result, err := r.db.Query(input)
	if err != nil {
		return models.LeaderBoard{}, err
	}

	userIDs := []models.UserID{}

	for _, item := range result.Items {
		userID, err := r.unmarshalLeaderboardItemFromDynamoDB(item)
		if err != nil {
			return models.LeaderBoard{}, err
		}
		userIDs = append(userIDs, userID)
	}

	return models.NewLeaderBoard(gameID, attr, userIDs), nil
}

func (r *DynamoDBLeaderboardRepository) GetBoundedLeaderboard(gameID models.GameID, attr models.AttributeName, limit int) (models.BoundedLeaderboard, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("Id = :id AND begins_with(#range, :attr)"),
		ExpressionAttributeNames: map[string]*string{
			"#range": aws.String("Range"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id":   {S: aws.String(fmt.Sprintf("Leaderboard.%s", gameID))},
			":attr": {S: aws.String(fmt.Sprintf("%s.", attr))},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int64(int64(limit)),
	}

	result, err := r.db.Query(input)
	if err != nil {
		return models.BoundedLeaderboard{}, err
	}

	userIDs := []models.UserID{}

	for _, item := range result.Items {
		userID, err := r.unmarshalLeaderboardItemFromDynamoDB(item)
		if err != nil {
			return models.BoundedLeaderboard{}, err
		}
		userIDs = append(userIDs, userID)
	}

	return models.NewBoundedLeaderBoard(gameID, attr, userIDs, limit), nil
}

func (r *DynamoDBLeaderboardRepository) AddLeaderboardItem(gameID models.GameID, userID models.UserID, attr models.AttributeName, value models.AttributeStat, tx *dynamodb.TransactWriteItemsInput) error {
	if !utils.AttributePositive(value) {
		return errors.New("attributes must be positive")
	}
	item, err := r.marshalLeaderboardItemToDynamoDB(gameID, attr, value, userID)
	if err != nil {
		return err
	}

	if tx != nil {
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{
			Put: &dynamodb.Put{
				TableName: aws.String(r.tableName),
				Item:      item,
			},
		})
		return nil
	}

	_, err = r.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})

	return err
}

func (r *DynamoDBLeaderboardRepository) UpdateLeaderboardItem(gameID models.GameID, userID models.UserID, attr models.AttributeName, value models.AttributeStat, oldValue models.AttributeStat, tx *dynamodb.TransactWriteItemsInput) error {
	// Create a new transaction if one wasn't provided
	localTx := tx
	if localTx == nil {
		localTx = &dynamodb.TransactWriteItemsInput{}
	}
	if oldValue == value {
		return nil
	}


	// Delete the old entry
	err := r.DeleteLeaderboardItem(gameID, userID, attr, oldValue, localTx)
	if err != nil {
		return err
	}

	// Add the new entry
	err = r.AddLeaderboardItem(gameID, userID, attr, value, localTx)
	if err != nil {
		return err
	}

	// If we created a new transaction, execute it
	if tx == nil {
		_, err = r.db.TransactWriteItems(localTx)
		return err
	}

	return nil
}

func (r *DynamoDBLeaderboardRepository) DeleteLeaderboardItem(gameID models.GameID, userID models.UserID, attr models.AttributeName, oldValue models.AttributeStat, tx *dynamodb.TransactWriteItemsInput) error {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id":    {S: aws.String(fmt.Sprintf("Leaderboard.%s", gameID))},
			"Range": {S: aws.String(fmt.Sprintf("%s.%05d.%s", attr, oldValue, userID))},
		},
	}

	result, err := r.db.GetItem(input)
	if err != nil {
		return err
	}

	if result.Item == nil {
		return nil // Item not found, nothing to delete
	}

	if tx != nil {
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{
			Delete: &dynamodb.Delete{
				TableName: aws.String(r.tableName),
				Key: map[string]*dynamodb.AttributeValue{
					"Id":    {S: aws.String(fmt.Sprintf("Leaderboard.%s", gameID))},
					"Range": {S: aws.String(fmt.Sprintf("%s.%05d.%s", attr, oldValue, userID))},
				},
			},
		})
		return nil
	}

	_, err = r.db.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id":    {S: aws.String(fmt.Sprintf("Leaderboard.%s", gameID))},
			"Range": {S: aws.String(fmt.Sprintf("%s.%05d.%s", attr, oldValue, userID))},
		},
	})

	return err
}

func (r *DynamoDBLeaderboardRepository) DeleteLeaderboardItemsByGameAndUser(gameID models.GameID, userID models.UserID, tx *dynamodb.TransactWriteItemsInput) error {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("Id = :id"),
		FilterExpression:       aws.String("UserId = :userID"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id":     {S: aws.String(fmt.Sprintf("Leaderboard.%s", gameID))},
			":userID": {S: aws.String(string(userID))},
		},
	}

	result, err := r.db.Query(input)
	if err != nil {
		return err
	}

	for _, item := range result.Items {
		if tx != nil {
			tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{
				Delete: &dynamodb.Delete{
					TableName: aws.String(r.tableName),
					Key: map[string]*dynamodb.AttributeValue{
						"Id":    item["Id"],
						"Range": item["Range"],
					},
				},
			})
		} else {
			_, err = r.db.DeleteItem(&dynamodb.DeleteItemInput{
				TableName: aws.String(r.tableName),
				Key: map[string]*dynamodb.AttributeValue{
					"Id":    item["Id"],
					"Range": item["Range"],
				},
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *DynamoDBLeaderboardRepository) DeleteLeaderboardItemsByGame(gameID models.GameID, tx *dynamodb.TransactWriteItemsInput) error {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("Id = :id"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id": {S: aws.String(fmt.Sprintf("Leaderboard.%s", gameID))},
		},
	}

	result, err := r.db.Query(input)
	if err != nil {
		return err
	}

	for _, item := range result.Items {
		if tx != nil {
			tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{
				Delete: &dynamodb.Delete{
					TableName: aws.String(r.tableName),
					Key: map[string]*dynamodb.AttributeValue{
						"Id":    item["Id"],
						"Range": item["Range"],
					},
				},
			})
		} else {
			_, err = r.db.DeleteItem(&dynamodb.DeleteItemInput{
				TableName: aws.String(r.tableName),
				Key: map[string]*dynamodb.AttributeValue{
					"Id":    item["Id"],
					"Range": item["Range"],
				},
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *DynamoDBLeaderboardRepository) DeleteLeaderboardItemsByGameAndAttribute(gameID models.GameID, attr models.AttributeName, tx *dynamodb.TransactWriteItemsInput) error {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("Id = :id AND begins_with(#range, :attr)"),
		ExpressionAttributeNames: map[string]*string{
			"#range": aws.String("Range"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id":   {S: aws.String(fmt.Sprintf("Leaderboard.%s", gameID))},
			":attr": {S: aws.String(fmt.Sprintf("%s.", attr))},
		},
	}

	result, err := r.db.Query(input)
	if err != nil {
		return err
	}

	for _, item := range result.Items {
		if tx != nil {
			tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{
				Delete: &dynamodb.Delete{
					TableName: aws.String(r.tableName),
					Key: map[string]*dynamodb.AttributeValue{
						"Id":    item["Id"],
						"Range": item["Range"],
					},
				},
			})
		} else {
			_, err = r.db.DeleteItem(&dynamodb.DeleteItemInput{
				TableName: aws.String(r.tableName),
				Key: map[string]*dynamodb.AttributeValue{
					"Id":    item["Id"],
					"Range": item["Range"],
				},
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *DynamoDBLeaderboardRepository) marshalLeaderboardItemToDynamoDB(gameID models.GameID, attr models.AttributeName, value models.AttributeStat, userID models.UserID) (map[string]*dynamodb.AttributeValue, error) {
	item := map[string]*dynamodb.AttributeValue{
		"Id":     {S: aws.String(fmt.Sprintf("Leaderboard.%s", gameID))},
		"Range":  {S: aws.String(fmt.Sprintf("%s.%05d.%s", attr, value, userID))},
		"UserId": {S: aws.String(string(userID))},
	}

	return item, nil
}

func (r *DynamoDBLeaderboardRepository) unmarshalLeaderboardItemFromDynamoDB(item map[string]*dynamodb.AttributeValue) (models.UserID, error) {
	var leaderboardItem struct {
		ID     string
		Range  string
		UserId string
	}

	err := dynamodbattribute.UnmarshalMap(item, &leaderboardItem)
	if err != nil {
		return "", err
	}

	return models.UserID(leaderboardItem.UserId), nil
}
