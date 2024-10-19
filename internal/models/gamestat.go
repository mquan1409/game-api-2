package models

import "errors"

type GameStat struct {
	UserID UserID `json:"UserID"`
	GameID GameID `json:"GameID"`
	GameAttributes AttributesStatsMap `json:"GameAttributes"`
}

func NewGameStat(userID UserID, gameID GameID, gameAttributes AttributesStatsMap) (*GameStat, error) {
	// Check if any attribute stat value is negative
	for _, value := range gameAttributes {
		if value < 0 {
			return nil, errors.New("attribute stat values cannot be negative")
		}
	}
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if gameID == "" {
		return nil, errors.New("game ID cannot be empty")
	}
	if gameAttributes == nil {
		return nil, errors.New("game attributes cannot be nil")
	}

	return &GameStat{
		UserID: userID,
		GameID: gameID,
		GameAttributes: gameAttributes,
	}, nil
}