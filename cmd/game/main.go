package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mquan1409/game-api/internal/config"
	"github.com/mquan1409/game-api/internal/handlers"
	"github.com/mquan1409/game-api/internal/repositories"
	"github.com/mquan1409/game-api/internal/services"
)

var gameHandler handlers.GameHandler

func init() {
	// Load configuration based on environment
	env := os.Getenv("APP_ENV")
	cfg := config.LoadConfig(env)

	// Create the session
	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(cfg.DynamoDBEndpoint),
		Region:   aws.String(cfg.DynamoDBRegion),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return
	}
	db := dynamodb.New(sess)

	// Initialize repository
	gameRepository := repositories.NewDynamoDBGameRepository(db, cfg.TableName)
	leaderboardRepository := repositories.NewDynamoDBLeaderboardRepository(db, cfg.TableName)
	gameService := services.NewGameServiceImpl(gameRepository, leaderboardRepository)

	// Initialize handler
	gameHandler = handlers.NewGameHandlerImpl(gameService)
}

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	path := req.Path
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	switch req.HTTPMethod {
	case "GET":
		if len(pathParts) == 2 && pathParts[0] == "games" {
			// GET /games/{id}
			return gameHandler.GetGame(req)
		} else if len(pathParts) == 4 && pathParts[0] == "games" && pathParts[2] == "leaderboard" {
			// GET /games/{gameId}/leaderboard/{attribute}
			if limit, ok := req.QueryStringParameters["limit"]; ok {
				// GET /games/{gameId}/leaderboard/{attribute}?limit={limit}
				req.QueryStringParameters["limit"] = limit
			}
			return gameHandler.GetLeaderboard(req)
		}
	case "POST":
		if len(pathParts) == 1 && pathParts[0] == "games" {
			// POST /games
			return gameHandler.CreateGame(req)
		}
	case "PUT":
		if len(pathParts) == 2 && pathParts[0] == "games" {
			// PUT /games/{id}
			return gameHandler.UpdateGame(req)
		}
	case "DELETE":
		if len(pathParts) == 2 && pathParts[0] == "games" {
			// DELETE /games/{id}
			return gameHandler.DeleteGame(req)
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 404,
		Body:       "Not Found",
	}, nil
}

func main() {
	lambda.Start(router)
}
