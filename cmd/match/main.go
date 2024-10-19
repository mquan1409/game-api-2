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

var matchHandler handlers.MatchHandler

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
	matchRepository := repositories.NewDynamoDBMatchRepository(db, cfg.TableName)
	gameRepository := repositories.NewDynamoDBGameRepository(db, cfg.TableName)
	gameStatRepository := repositories.NewDynamoDBGameStatRepository(db, cfg.TableName)
	leaderboardRepository := repositories.NewDynamoDBLeaderboardRepository(db, cfg.TableName)
	matchService := services.NewMatchServiceImpl(matchRepository, gameRepository, gameStatRepository, leaderboardRepository)

	// Initialize handler
	matchHandler = handlers.NewMatchHandlerImpl(matchService)
}

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	path := req.Path
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	switch req.HTTPMethod {
	case "GET":
		if len(pathParts) == 4 && pathParts[0] == "matches" {
			// GET /matches/{gameId}/{matchId}/{dateId}
			return matchHandler.GetMatch(req)
		} else if len(pathParts) == 1 && pathParts[0] == "matches" {
			// GET /matches?game={gameId}&date={dateId}
			return matchHandler.GetMatchesByGameAndDate(req)
		}
	case "POST":
		if len(pathParts) == 1 && pathParts[0] == "matches" {
			// POST /matches
			return matchHandler.CreateMatch(req)
		}
	case "PUT":
		if len(pathParts) == 2 && pathParts[0] == "matches" {
			// PUT /matches/{id}
			return matchHandler.UpdateMatch(req)
		}
	case "DELETE":
		if len(pathParts) == 4 && pathParts[0] == "matches" {
			// DELETE /matches/{gameId}/{matchId}/{dateId}
			return matchHandler.DeleteMatch(req)
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
