package repositories

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/mquan1409/game-api/internal/models"
)

type DynamoDBUserRepository struct {
	db *dynamodb.DynamoDB
	tableName string
}

func NewDynamoDBUserRepository(db *dynamodb.DynamoDB, tableName string) UserRepository {
	return &DynamoDBUserRepository{db: db, tableName: tableName}
}

func (r *DynamoDBUserRepository) GetUser(id models.UserID) (*models.User, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String("USER_INFO-prefix:" + string(id[:1])),
			},
			"Range": {
				S: aws.String(string(id)),
			},
		},
	}

	result, err := r.db.GetItem(input)
	
	if err != nil {
		fmt.Printf("Error getting item from DynamoDB: %v\n", err)
		return nil, err
	}

	if result.Item == nil {
		fmt.Printf("User not found in DynamoDB\n")
		return nil, errors.New("user not found")
	}

	fmt.Printf("Successfully retrieved user from DynamoDB\n")

	return r.unmarshalUserFromDynamoDB(result.Item)
}

func (r *DynamoDBUserRepository) GetUserBasicsByPrefix(prefix string) ([]*models.UserBasic, error) {
	if prefix == "" {
		return nil, errors.New("prefix cannot be empty")
	}

	keyCondition := expression.Key("Id").Equal(expression.Value("USER_INFO-prefix:" + string(prefix[:1])))
	keyCondition = keyCondition.And(expression.Key("Range").BeginsWith(prefix))

	proj := expression.NamesList(expression.Name("Id"), expression.Name("Range"), expression.Name("Username"))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).WithProjection(proj).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.tableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
	}

	result, err := r.db.Query(input)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	var userBasics []*models.UserBasic
	for _, item := range result.Items {
		userBasic, err := r.unmarshalUserBasicFromDynamoDB(item)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal user basic: %w", err)
		}
		userBasics = append(userBasics, userBasic)
	}

	return userBasics, nil
}

func (r *DynamoDBUserRepository) CreateUser(user *models.User, tx *dynamodb.TransactWriteItemsInput) (*models.User, error) {
	if user.UserID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	item, err := r.marshalUserToDynamoDBAttributeValue(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user: %w", err)
	}

	putItem := &dynamodb.Put{
		TableName: aws.String(r.tableName),
		Item:      item,
	}

	if tx != nil {
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{
			Put: putItem,
		})
		return user, nil
	}

	_, err = r.db.PutItem(&dynamodb.PutItemInput{
		Item:      putItem.Item,
		TableName: putItem.TableName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *DynamoDBUserRepository) UpdateUser(user *models.User, tx *dynamodb.TransactWriteItemsInput) (*models.User, error) {
	if user.UserID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	item, err := r.marshalUserToDynamoDBAttributeValue(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user: %w", err)
	}

	putItem := &dynamodb.Put{
		TableName: aws.String(r.tableName),
		Item:      item,
	}

	if tx != nil {
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{
			Put: putItem,
		})
		return user, nil
	}

	_, err = r.db.PutItem(&dynamodb.PutItemInput{
		Item:      putItem.Item,
		TableName: putItem.TableName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (r *DynamoDBUserRepository) DeleteUser(id *models.UserID, tx *dynamodb.TransactWriteItemsInput) error {
	if *id == "" {
		return errors.New("id cannot be empty")
	}

	deleteInput := r.createDeleteInput(*id)

	if tx != nil {
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{
			Delete: deleteInput,
		})
		return nil
	}

	_, err := r.db.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: deleteInput.TableName,
		Key:       deleteInput.Key,
	})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (r *DynamoDBUserRepository) createDeleteInput(id models.UserID) *dynamodb.Delete {
	return &dynamodb.Delete{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String("USER_INFO-prefix:" + string(id[:1])),
			},
			"Range": {
				S: aws.String(string(id)),
			},
		},
	}
}

func (r *DynamoDBUserRepository) marshalUserToDynamoDBAttributeValue(user *models.User) (map[string]*dynamodb.AttributeValue, error) {
	av := make(map[string]*dynamodb.AttributeValue)
	av["Id"] = &dynamodb.AttributeValue{S: aws.String(fmt.Sprintf("USER_INFO-prefix:%s", string(user.UserID[:1])))}
	av["Range"] = &dynamodb.AttributeValue{S: aws.String(string(user.UserID))}
	av["Username"] = &dynamodb.AttributeValue{S: aws.String(user.Username)}
	av["GamesPlayed"] = &dynamodb.AttributeValue{L: make([]*dynamodb.AttributeValue, len(user.GamesPlayed))}

	for i, gameID := range user.GamesPlayed {
		av["GamesPlayed"].L[i] = &dynamodb.AttributeValue{S: aws.String(string(gameID))}
	}
	av["Email"] = &dynamodb.AttributeValue{S: aws.String(user.Email)}

	return av, nil
}

func (r *DynamoDBUserRepository) unmarshalUserFromDynamoDB(item map[string]*dynamodb.AttributeValue) (*models.User, error) {
	var userID models.UserID
	var username, email string
	var gamesPlayed []models.GameID

	// Unmarshal UserID
	if idAttr, ok := item["Range"]; ok && idAttr.S != nil {
		userID = models.UserID(*idAttr.S)
	} else {
		return nil, errors.New("error UserID is missing or invalid")
	}

	// Unmarshal Username
	if usernameAttr, ok := item["Username"]; ok && usernameAttr.S != nil {
		username = *usernameAttr.S
	} else {
		return nil, errors.New("error Username is missing or invalid")
	}

	// Unmarshal Email
	if emailAttr, ok := item["Email"]; ok && emailAttr.S != nil {
		email = *emailAttr.S
	} else {
		return nil, errors.New("error Email is missing or invalid")
	}

	// Unmarshal GamesPlayed
	if gamesPlayedAttr, ok := item["GamesPlayed"]; ok && gamesPlayedAttr.L != nil {
		for _, gameIDAttr := range gamesPlayedAttr.L {
			if gameIDAttr.S != nil {
				gamesPlayed = append(gamesPlayed, models.GameID(*gameIDAttr.S))
			}
		}
	}

	return models.NewUser(userID, username, email, gamesPlayed)
}

func (r *DynamoDBUserRepository) unmarshalUserBasicFromDynamoDB(item map[string]*dynamodb.AttributeValue) (*models.UserBasic, error) {
	var userID models.UserID
	var username string

	// Unmarshal UserID
	if idAttr, ok := item["Range"]; ok && idAttr.S != nil {
		userID = models.UserID(*idAttr.S)
	} else {
		return nil, errors.New("error UserID is missing or invalid")
	}

	// Unmarshal Username
	if usernameAttr, ok := item["Username"]; ok && usernameAttr.S != nil {
		username = *usernameAttr.S
	} else {
		return nil, errors.New("error Username is missing or invalid")
	}

	userBasic, err := models.NewUserBasic(userID, username)
	if err != nil {
		return nil, fmt.Errorf("failed to create UserBasic: %w", err)
	}
	return userBasic, nil
}
