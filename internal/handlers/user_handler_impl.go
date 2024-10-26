package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mquan1409/game-api/internal/models"	
	"github.com/mquan1409/game-api/internal/services"
)

type UserHandlerImpl struct {
	userService services.UserService
}

func NewUserHandlerImpl(userService services.UserService) UserHandler {
	return &UserHandlerImpl{userService: userService}
}

func (h *UserHandlerImpl) GetUser(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID := models.UserID(event.PathParameters["userId"])
	user, err := h.userService.GetUser(userID)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}
	
	userJSON, err := json.Marshal(user)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal user data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(userJSON),
	}, nil
}

func (h *UserHandlerImpl) GetUserBasicsByPrefix(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	prefix := event.QueryStringParameters["prefix"]
	users, err := h.userService.GetUserBasicsByPrefix(prefix)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	usersJSON, err := json.Marshal(users)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal users data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(usersJSON),
	}, nil
}

func (h *UserHandlerImpl) GetGameStat(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID := models.UserID(event.PathParameters["userId"])
	gameID := models.GameID(event.PathParameters["gameId"])
	
	gameStat, err := h.userService.GetGameStat(userID, gameID)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	gameStatJSON, err := json.Marshal(gameStat)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal game stat data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(gameStatJSON),
	}, nil
}

func (h *UserHandlerImpl) CreateUser(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var user models.User
	err := json.Unmarshal([]byte(event.Body), &user)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request body",
		}, nil
	}

	createdUser, err := h.userService.CreateUser(&user)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	createdUserJSON, err := json.Marshal(createdUser)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal created user data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Body:       string(createdUserJSON),
	}, nil
}

func (h *UserHandlerImpl) UpdateUser(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var user models.User
	err := json.Unmarshal([]byte(event.Body), &user)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request body",
		}, nil
	}

	updatedUser, err := h.userService.UpdateUser(&user)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	updatedUserJSON, err := json.Marshal(updatedUser)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal updated user data",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(updatedUserJSON),
	}, nil
}

func (h *UserHandlerImpl) DeleteUser(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID := models.UserID(event.PathParameters["userId"])
	err := h.userService.DeleteUser(&userID)
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