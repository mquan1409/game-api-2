package handlers

import (
	"github.com/aws/aws-lambda-go/events"
)

type MatchHandler interface {
	GetMatch(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	GetMatchesByGameAndDate(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	CreateMatch(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	UpdateMatch(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	DeleteMatch(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
}
