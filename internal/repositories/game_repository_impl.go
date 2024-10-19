package repositories

import (
	"fmt"
	"github.com/mquan1409/game-api/internal/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"errors"
)

type DynamoDBGameRepository struct {
	db *dynamodb.DynamoDB
	tableName string
}

func NewDynamoDBGameRepository(db *dynamodb.DynamoDB, tableName string) GameRepository {
	return &DynamoDBGameRepository{db: db, tableName: tableName}
}

func (r *DynamoDBGameRepository) GetGame(id models.GameID) (*models.Game, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String("GAME_INFO"),
			},
			"Range": {
				S: aws.String(string(id)),
			},
		},
	}

	result, err := r.db.GetItem(input)
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, errors.New("game not found")
	}

	game, err := unmarshalGameFromDynamoDB(result.Item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal game: %w", err)
	}

	return game, nil
}

func (r *DynamoDBGameRepository) CreateGame(game *models.Game, tx *dynamodb.TransactWriteItemsInput) (*models.Game, error) {
	item, err := r.marshalGameToDynamoDBAttributeValue(game)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal game: %w", err)
	}

	putItem := &dynamodb.Put{
		Item:      item,
		TableName: aws.String(r.tableName),
	}

	if tx != nil {
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{
			Put: putItem,
		})
		return game, nil
	}	

	input := &dynamodb.PutItemInput{
		Item:      putItem.Item,
		TableName: putItem.TableName,
	}

	_, err = r.db.PutItem(input)
	if err != nil {
		return nil, err
	}

	return game, nil
}

func (r *DynamoDBGameRepository) UpdateGame(game *models.Game, tx *dynamodb.TransactWriteItemsInput) (*models.Game, error) {
	item, err := r.marshalGameToDynamoDBAttributeValue(game)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal game: %w", err)
	}

	putItem := &dynamodb.Put{
		Item:      item,
		TableName: aws.String(r.tableName),
	}

	if tx != nil {
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{
			Put: putItem,
		})
		return game, nil
	}

	input := &dynamodb.PutItemInput{
		Item:      putItem.Item,
		TableName: putItem.TableName,
	}

	_, err = r.db.PutItem(input)
	if err != nil {
		return nil, err
	}

	return game, nil
}

func (r *DynamoDBGameRepository) DeleteGame(id models.GameID, tx *dynamodb.TransactWriteItemsInput) error {
	deleteItem := &dynamodb.Delete{
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String("GAME_INFO"),
			},
			"Range": {
				S: aws.String(string(id)),
			},
		},
		TableName: aws.String(r.tableName),
	}

	if tx != nil {
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{
			Delete: deleteItem,
		})
		return nil
	}

	_, err := r.db.DeleteItem(&dynamodb.DeleteItemInput{
		Key:       deleteItem.Key,
		TableName: deleteItem.TableName,
	})
	return err
}

func unmarshalGameFromDynamoDB(item map[string]*dynamodb.AttributeValue) (*models.Game, error) {
	if item == nil {
		return nil, errors.New("item is nil")
	}

	gameID := models.GameID(*item["Range"].S)
	description := *item["Description"].S

	var attributes []models.AttributeName
	if attributesAV, ok := item["Attributes"]; ok && attributesAV.L != nil {
		for _, av := range attributesAV.L {
			if av.S != nil {
				attributes = append(attributes, models.AttributeName(*av.S))
			}
		}
	}

	var rankedAttributes []models.AttributeName
	if rankedAttributesAV, ok := item["RankedAttributes"]; ok && rankedAttributesAV.L != nil {
		for _, av := range rankedAttributesAV.L {
			if av.S != nil {
				rankedAttributes = append(rankedAttributes, models.AttributeName(*av.S))
			}
		}
	}

	return models.NewGame(gameID, description, attributes, rankedAttributes)
}

func (r *DynamoDBGameRepository) marshalGameToDynamoDBAttributeValue(game *models.Game) (map[string]*dynamodb.AttributeValue, error) {
	av := make(map[string]*dynamodb.AttributeValue)
	av["Id"] = &dynamodb.AttributeValue{S: aws.String("GAME_INFO")}
	av["Range"] = &dynamodb.AttributeValue{S: aws.String(string(game.GameID))}
	av["Description"] = &dynamodb.AttributeValue{S: aws.String(game.Description)}
	av["Attributes"] = &dynamodb.AttributeValue{L: make([]*dynamodb.AttributeValue, len(game.Attributes))}

	for i, attr := range game.Attributes {
		av["Attributes"].L[i] = &dynamodb.AttributeValue{S: aws.String(string(attr))}
	}

	av["RankedAttributes"] = &dynamodb.AttributeValue{L: make([]*dynamodb.AttributeValue, len(game.RankedAttributes))}

	for i, attr := range game.RankedAttributes {
		av["RankedAttributes"].L[i] = &dynamodb.AttributeValue{S: aws.String(string(attr))}
	}

	return av, nil
}
