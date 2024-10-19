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

var userHandler handlers.UserHandler

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
	userRepository := repositories.NewDynamoDBUserRepository(db, cfg.TableName)
	gameStatRepository := repositories.NewDynamoDBGameStatRepository(db, cfg.TableName)
	userService := services.NewUserServiceImpl(userRepository, gameStatRepository)

	// Initialize handler
	userHandler = handlers.NewUserHandlerImpl(userService)
}

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	path := req.Path
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	switch req.HTTPMethod {
	case "GET":
		if len(pathParts) == 2 && pathParts[0] == "users" {
			// GET /users/{id}
			return userHandler.GetUser(req)
		} else if len(pathParts) == 1 && pathParts[0] == "users" && req.QueryStringParameters["prefix"] != "" {
			// GET /users?prefix={prefix}
			return userHandler.GetUserBasicsByPrefix(req)
		} else if len(pathParts) == 5 && pathParts[0] == "users" && pathParts[2] == "games" && pathParts[4] == "stats" {
			// GET /users/{userId}/games/{gameId}/stats
			return userHandler.GetGameStat(req)
		}
	case "POST":
		if len(pathParts) == 1 && pathParts[0] == "users" {
			// POST /users
			return userHandler.CreateUser(req)
		}
	case "PUT":
		if len(pathParts) == 2 && pathParts[0] == "users" {
			// PUT /users/{id}
			return userHandler.UpdateUser(req)
		}
	case "DELETE":
		if len(pathParts) == 2 && pathParts[0] == "users" {
			// DELETE /users/{id}
			return userHandler.DeleteUser(req)
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
