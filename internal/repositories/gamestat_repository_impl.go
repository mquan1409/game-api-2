package repositories

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mquan1409/game-api/internal/models"
	"github.com/mquan1409/game-api/internal/utils"
)

type GameStatDynamoDBRepository struct {
	db        *dynamodb.DynamoDB
	tableName string
}

func NewDynamoDBGameStatRepository(db *dynamodb.DynamoDB, tableName string) GameStatRepository {
	return &GameStatDynamoDBRepository{db: db, tableName: tableName}
}

func (r *GameStatDynamoDBRepository) GetGameStat(userID models.UserID, gameID models.GameID) (*models.GameStat, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id":    {S: aws.String(fmt.Sprintf("GameStat.%s", userID))},
			"Range": {S: aws.String(string(gameID))},
		},
	}

	result, err := r.db.GetItem(input)
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, errors.New("game stat not found")
	}

	return r.unmarshalGameStatFromDynamoDB(result.Item)
}

func (r *GameStatDynamoDBRepository) CreateGameStat(gameStat *models.GameStat, tx *dynamodb.TransactWriteItemsInput) error {
	if !utils.AttributesPositive(gameStat.GameAttributes) {
		return errors.New("attributes must be positive")
	}
	av, err := r.marshalGameStatToDynamoDBAttributeValue(gameStat)
	if err != nil {
		return err
	}

	input := &dynamodb.Put{
		TableName: aws.String(r.tableName),
		Item:      av,
	}
	if tx != nil {
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{Put: input})
	}

	_, err = r.db.PutItem(&dynamodb.PutItemInput{
		TableName: input.TableName,
		Item:      input.Item,
	})
	return err
}

func (r *GameStatDynamoDBRepository) UpdateGameStat(gameStat *models.GameStat, tx *dynamodb.TransactWriteItemsInput) error {
	if !utils.AttributesPositive(gameStat.GameAttributes) {
		return errors.New("attributes must be positive")
	}
	av, err := r.marshalGameStatToDynamoDBAttributeValue(gameStat)
	if err != nil {
		return err
	}

	input := &dynamodb.Put{
		TableName: aws.String(r.tableName),
		Item:      av,
	}
	if tx != nil {
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{Put: input})
	}

	_, err = r.db.PutItem(&dynamodb.PutItemInput{
		TableName: input.TableName,
		Item:      input.Item,
	})
	return err
}

func (r *GameStatDynamoDBRepository) DeleteGameStat(userID models.UserID, gameID models.GameID, tx *dynamodb.TransactWriteItemsInput) error {
	input := &dynamodb.Delete{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id":    {S: aws.String(fmt.Sprintf("GameStat.%s", userID))},
			"Range": {S: aws.String(string(gameID))},
		},
	}
	if tx != nil {	
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{Delete: input})
	}

	_, err := r.db.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: input.TableName,
		Key:       input.Key,
	})
	return err
}

func (r *GameStatDynamoDBRepository) unmarshalGameStatFromDynamoDB(item map[string]*dynamodb.AttributeValue) (*models.GameStat, error) {
	// Extract UserID and GameID
	idParts := strings.SplitN(*item["Id"].S, ".", 2)
	if len(idParts) != 2 {
		return nil, errors.New("invalid Id format")
	}
	userID := models.UserID(idParts[1])
	gameID := models.GameID(*item["Range"].S)

	// Extract GameAttributes
	gameAttributes := make(models.AttributesStatsMap)
	for attrName, attrValue := range item["Attributes"].M {
		value, err := strconv.Atoi(*attrValue.N)
		if err != nil {
			return nil, err
		}
		gameAttributes[models.AttributeName(attrName)] = models.AttributeStat(value)
	}

	return models.NewGameStat(userID, gameID, gameAttributes)
}

func (r *GameStatDynamoDBRepository) marshalGameStatToDynamoDBAttributeValue(gameStat *models.GameStat) (map[string]*dynamodb.AttributeValue, error) {
	av := make(map[string]*dynamodb.AttributeValue)

	// Set Id and Range
	av["Id"] = &dynamodb.AttributeValue{S: aws.String(fmt.Sprintf("GameStat.%s", gameStat.UserID))}
	av["Range"] = &dynamodb.AttributeValue{S: aws.String(string(gameStat.GameID))}

	// Set Attributes
	attributes := make(map[string]*dynamodb.AttributeValue)
	for attrName, attrValue := range gameStat.GameAttributes {
		attributes[string(attrName)] = &dynamodb.AttributeValue{N: aws.String(strconv.Itoa(int(attrValue)))}
	}
	av["Attributes"] = &dynamodb.AttributeValue{M: attributes}

	return av, nil
}
