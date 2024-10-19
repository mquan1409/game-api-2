package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func ScanEntireTable(db *dynamodb.DynamoDB, tableName string) ([]map[string]*dynamodb.AttributeValue, error) {
	var results []map[string]*dynamodb.AttributeValue
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	for {
		output, err := db.Scan(input)
		if err != nil {
			return nil, err
		}
		results = append(results, output.Items...)

		if output.LastEvaluatedKey == nil {
			break
		}
		input.ExclusiveStartKey = output.LastEvaluatedKey
	}

	return results, nil
}