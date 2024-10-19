package repositories

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mquan1409/game-api/internal/models"
	"errors"
	"fmt"
	"strconv"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type MatchDynamoDBRepository struct {
	db *dynamodb.DynamoDB
	tableName string
}

func NewDynamoDBMatchRepository(db *dynamodb.DynamoDB, tableName string) MatchRepository {
	return &MatchDynamoDBRepository{db: db, tableName: tableName}
}


func (r *MatchDynamoDBRepository) GetMatch(gameID models.GameID, matchID models.MatchID, dateID models.DateID) (*models.Match, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(fmt.Sprintf("MATCH_INFO.%s", gameID)),
			},
			"Range": {
				S: aws.String(fmt.Sprintf("%s.%s", dateID, matchID)),
			},
		},
	}

	result, err := r.db.GetItem(input)
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, errors.New("match not found")
	}

	match, err := r.unmarshalMatchFromDynamoDB(result.Item)
	if err != nil {
		return nil, err
	}

	return match, nil
}

func (r *MatchDynamoDBRepository) GetMatchesByGameAndDate(gameID models.GameID, dateID models.DateID) ([]*models.Match, error) {
	input := &dynamodb.QueryInput{
		TableName: aws.String(r.tableName),
		KeyConditionExpression: aws.String("Id = :gameID AND begins_with(#range, :dateID)"),
		ExpressionAttributeNames: map[string]*string{
			"#range": aws.String("Range"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gameID": {S: aws.String(fmt.Sprintf("MATCH_INFO.%s", gameID))},
			":dateID": {S: aws.String(fmt.Sprintf("%s.", dateID))},
		},
	}

	result, err := r.db.Query(input)
	if err != nil {
		return nil, err
	}

	var matches []*models.Match
	for _, item := range result.Items {
		match, err := r.unmarshalMatchFromDynamoDB(item)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}

	return matches, nil
}


func (r *MatchDynamoDBRepository) CreateMatch(match *models.Match, tx *dynamodb.TransactWriteItemsInput) (*models.Match, error) {
	av, err := r.marshalMatchToDynamoDBAttributeValue(match)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.Put{
		Item:      av,
		TableName: aws.String(r.tableName),
	}
	if tx != nil {
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{Put: input})
		return match, nil
	}

	_, err = r.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	return match, err
}

func (r *MatchDynamoDBRepository) UpdateMatch(match *models.Match, tx *dynamodb.TransactWriteItemsInput) (*models.Match, error) {
	av, err := r.marshalMatchToDynamoDBAttributeValue(match)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.Put{
		Item:      av,
		TableName: aws.String(r.tableName),
	}
	if tx != nil {
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{Put: input})
		return match, nil
	}

	_, err = r.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	return match, err
}

func (r *MatchDynamoDBRepository) DeleteMatch(gameID models.GameID, matchID models.MatchID, dateID models.DateID, tx *dynamodb.TransactWriteItemsInput) error {
	input := &dynamodb.Delete{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(fmt.Sprintf("MATCH_INFO.%s", string(gameID))),
			},
			"Range": {
				S: aws.String(fmt.Sprintf("%s.%s", string(dateID), string(matchID))),
			},
		},
	}
	if tx != nil {
		tx.TransactItems = append(tx.TransactItems, &dynamodb.TransactWriteItem{Delete: input})
		return nil
	}

	_, err := r.db.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: input.TableName,
		Key:       input.Key,
	})
	return err
}

func (r *MatchDynamoDBRepository) unmarshalMatchFromDynamoDB(item map[string]*dynamodb.AttributeValue) (*models.Match, error) {
	// Unmarshal basic fields
	matchID := models.MatchID((*item["Range"].S)[11:])
	dateID := models.DateID((*item["Range"].S)[:10])
	gameID := models.GameID((*item["Id"].S)[11:])

	// Unmarshal TeamNames
	var teamNames []string
	for _, nameAV := range item["TeamNames"].L {
		teamNames = append(teamNames, *nameAV.S)
	}

	// Unmarshal TeamScores
	var teamScores []int
	for _, scoreAV := range item["TeamScores"].L {
		score, err := strconv.Atoi(*scoreAV.N)
		if err != nil {
			return nil, err
		}
		teamScores = append(teamScores, score)
	}

	// Unmarshal TeamMembers
	var teamMembers [][]string
	for _, teamAV := range item["TeamMembers"].L {
		var team []string
		for _, memberAV := range teamAV.L {
			team = append(team, *memberAV.S)
		}
		teamMembers = append(teamMembers, team)
	}

	// Unmarshal PlayerAttributes
	playerAttributesMap := make(map[models.UserID]models.AttributesStatsMap)
	for userID, attributesAV := range item["PlayerAttributes"].M {
		attributesMap := make(models.AttributesStatsMap)
		for attrName, attrValueAV := range attributesAV.M {
			attrValue, err := strconv.Atoi(*attrValueAV.N)
			if err != nil {
				return nil, err
			}
			attributesMap[models.AttributeName(attrName)] = models.AttributeStat(attrValue)
		}
		playerAttributesMap[models.UserID(userID)] = attributesMap
	}

	// Use NewMatch constructor without playerAttributesMap
	match, err := models.NewMatch(matchID, dateID, gameID, teamNames, teamScores, teamMembers, playerAttributesMap)
	if err != nil {
		return nil, err
	}

	return match, nil
}

func (r *MatchDynamoDBRepository) marshalMatchToDynamoDBAttributeValue(match *models.Match) (map[string]*dynamodb.AttributeValue, error) {
	av := make(map[string]*dynamodb.AttributeValue)

	// Set Id and Range
	av["Id"] = &dynamodb.AttributeValue{S: aws.String(fmt.Sprintf("MATCH_INFO.%s", match.GameID))}
	av["Range"] = &dynamodb.AttributeValue{S: aws.String(fmt.Sprintf("%s.%s", match.DateID, match.MatchID))}

	// Marshal TeamNames
	teamNames, err := dynamodbattribute.MarshalList(match.TeamNames)
	if err != nil {
		return nil, err
	}
	av["TeamNames"] = &dynamodb.AttributeValue{L: teamNames}

	// Marshal TeamScores
	teamScores := make([]*dynamodb.AttributeValue, len(match.TeamScores))
	for i, score := range match.TeamScores {
		teamScores[i] = &dynamodb.AttributeValue{N: aws.String(strconv.Itoa(score))}
	}
	av["TeamScores"] = &dynamodb.AttributeValue{L: teamScores}

	// Marshal TeamMembers
	teamMembers := make([]*dynamodb.AttributeValue, len(match.TeamMembers))
	for i, team := range match.TeamMembers {
		teamAV, err := dynamodbattribute.MarshalList(team)
		if err != nil {
			return nil, err
		}
		teamMembers[i] = &dynamodb.AttributeValue{L: teamAV}
	}
	av["TeamMembers"] = &dynamodb.AttributeValue{L: teamMembers}

	// Marshal PlayerAttributes
	playerAttrs := make(map[string]*dynamodb.AttributeValue)
	for userID, attrs := range match.PlayerAttributesMap {
		userAttrs := make(map[string]*dynamodb.AttributeValue)
		for attrName, attrValue := range attrs {
			userAttrs[string(attrName)] = &dynamodb.AttributeValue{N: aws.String(strconv.Itoa(int(attrValue)))}
		}
		playerAttrs[string(userID)] = &dynamodb.AttributeValue{M: userAttrs}
	}
	av["PlayerAttributes"] = &dynamodb.AttributeValue{M: playerAttrs}

	return av, nil
}
