package handlers

import (
	"github.com/aws/aws-lambda-go/events"
)

type GameHandler interface {
	GetGame(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	GetLeaderboard(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	GetBoundedLeaderboard(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	CreateGame(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	UpdateGame(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	DeleteGame(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
}
