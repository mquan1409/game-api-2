package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mquan1409/game-api/internal/models"
	"github.com/mquan1409/game-api/internal/services"
)

type MatchHandlerImpl struct {
	matchService services.MatchService
}

func NewMatchHandlerImpl(matchService services.MatchService) MatchHandler {
	return &MatchHandlerImpl{
		matchService: matchService,
	}
}

func (h *MatchHandlerImpl) GetMatch(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	gameID := models.GameID(event.PathParameters["gameId"])
	matchID := models.MatchID(event.PathParameters["matchId"])
	dateID := models.DateID(event.PathParameters["dateId"])

	match, err := h.matchService.GetMatch(gameID, matchID, dateID)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	matchJSON, err := json.Marshal(match)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal match data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(matchJSON),
	}, nil
}

func (h *MatchHandlerImpl) GetMatchesByGameAndDate(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	gameID := models.GameID(event.QueryStringParameters["game"])
	dateID := models.DateID(event.QueryStringParameters["date"])

	matches, err := h.matchService.GetMatchesByGameAndDate(gameID, dateID)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	matchesJSON, err := json.Marshal(matches)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal matches data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(matchesJSON),
	}, nil
}

func (h *MatchHandlerImpl) CreateMatch(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var match models.Match
	err := json.Unmarshal([]byte(event.Body), &match)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request body",
		}, nil
	}

	createdMatch, err := h.matchService.CreateMatch(&match)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	createdMatchJSON, err := json.Marshal(createdMatch)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal created match data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Body:       string(createdMatchJSON),
	}, nil
}

func (h *MatchHandlerImpl) UpdateMatch(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var match models.Match
	err := json.Unmarshal([]byte(event.Body), &match)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request body",
		}, nil
	}

	updatedMatch, err := h.matchService.UpdateMatch(&match)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	updatedMatchJSON, err := json.Marshal(updatedMatch)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal updated match data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(updatedMatchJSON),
	}, nil
}

func (h *MatchHandlerImpl) DeleteMatch(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	gameID := models.GameID(event.PathParameters["gameId"])
	matchID := models.MatchID(event.PathParameters["matchId"])
	dateID := models.DateID(event.PathParameters["dateId"])

	err := h.matchService.DeleteMatch(gameID, matchID, dateID)
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