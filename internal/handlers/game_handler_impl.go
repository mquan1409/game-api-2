package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mquan1409/game-api/internal/models"
	"github.com/mquan1409/game-api/internal/services"
)

type GameHandlerImpl struct {
	gameService services.GameService
}

func NewGameHandlerImpl(gameService services.GameService) GameHandler {
	return &GameHandlerImpl{
		gameService: gameService,
	}
}

func (h *GameHandlerImpl) GetGame(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	gameID := models.GameID(event.PathParameters["gameId"])
	game, err := h.gameService.GetGame(gameID)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	gameJSON, err := json.Marshal(game)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal game data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(gameJSON),
	}, nil
}

func (h *GameHandlerImpl) GetLeaderboard(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	gameID := models.GameID(event.PathParameters["gameId"])
	attribute := models.AttributeName(event.PathParameters["attribute"])
	var leaderboard *models.LeaderBoard
	var err error
	leaderboard, err = h.gameService.GetGameLeaderboard(gameID, attribute)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	leaderboardJSON, err := json.Marshal(leaderboard)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal leaderboard data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(leaderboardJSON),
	}, nil
}

func (h *GameHandlerImpl) GetBoundedLeaderboard(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	gameID := models.GameID(event.PathParameters["gameId"])
	attribute := models.AttributeName(event.PathParameters["attribute"])
	limit, _ := strconv.Atoi(event.QueryStringParameters["limit"])
	var leaderboard *models.BoundedLeaderboard
	var err error
	leaderboard, err = h.gameService.GetBoundedGameLeaderboard(gameID, attribute, limit)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	leaderboardJSON, err := json.Marshal(leaderboard)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal leaderboard data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(leaderboardJSON),
	}, nil
}
func (h *GameHandlerImpl) CreateGame(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var game models.Game
	err := json.Unmarshal([]byte(event.Body), &game)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request body",
		}, nil
	}

	createdGame, err := h.gameService.CreateGame(&game)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	createdGameJSON, err := json.Marshal(createdGame)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal created game data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Body:       string(createdGameJSON),
	}, nil
}

func (h *GameHandlerImpl) UpdateGame(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var game models.Game
	err := json.Unmarshal([]byte(event.Body), &game)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request body",
		}, nil
	}

	game.GameID = models.GameID(event.PathParameters["gameId"])

	updatedGame, err := h.gameService.UpdateGame(&game)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	updatedGameJSON, err := json.Marshal(updatedGame)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal updated game data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(updatedGameJSON),
	}, nil
}

func (h *GameHandlerImpl) DeleteGame(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	gameID := models.GameID(event.PathParameters["gameId"])

	err := h.gameService.DeleteGame(gameID)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
	}, nil
}