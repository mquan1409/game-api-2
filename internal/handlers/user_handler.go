package handlers

import (
	"github.com/aws/aws-lambda-go/events"
)

type UserHandler interface {
	GetUser(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	GetUserBasicsByPrefix(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	GetGameStat(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	CreateUser(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	UpdateUser(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	DeleteUser(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
}
